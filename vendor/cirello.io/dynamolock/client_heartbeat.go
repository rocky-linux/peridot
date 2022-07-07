/*
Copyright 2019 github.com/ucirello

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dynamolock

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

// SendHeartbeatOption allows to proceed with Lock content changes in the
// heartbeat cycle.
type SendHeartbeatOption func(*sendHeartbeatOptions)

type sendHeartbeatOptions struct {
	lockItem   *Lock
	data       []byte
	deleteData bool
}

// DeleteData removes the Lock data on heartbeat.
func DeleteData() SendHeartbeatOption {
	return func(o *sendHeartbeatOptions) {
		o.deleteData = true
	}
}

// ReplaceHeartbeatData overrides the content of the Lock in the heartbeat cycle.
func ReplaceHeartbeatData(data []byte) SendHeartbeatOption {
	return func(o *sendHeartbeatOptions) {
		o.deleteData = false
		o.data = data
	}
}

// SendHeartbeat indicates that the given lock is still being worked on. If
// using WithHeartbeatPeriod > 0 when setting up this object, then this method
// is unnecessary, because the background thread will be periodically calling it
// and sending heartbeats. However, if WithHeartbeatPeriod = 0, then this method
// must be called to instruct DynamoDB that the lock should not be expired.
func (c *Client) SendHeartbeat(lockItem *Lock, opts ...SendHeartbeatOption) error {
	return c.SendHeartbeatWithContext(context.Background(), lockItem, opts...)
}

// SendHeartbeatWithContext indicates that the given lock is still being worked
// on. If using WithHeartbeatPeriod > 0 when setting up this object, then this
// method is unnecessary, because the background thread will be periodically
// calling it and sending heartbeats. However, if WithHeartbeatPeriod = 0, then
// this method must be called to instruct DynamoDB that the lock should not be
// expired. The given context is passed down to the underlying dynamoDB call.
func (c *Client) SendHeartbeatWithContext(ctx context.Context, lockItem *Lock, opts ...SendHeartbeatOption) error {
	if c.isClosed() {
		return ErrClientClosed
	}
	sho := &sendHeartbeatOptions{
		lockItem: lockItem,
	}
	for _, opt := range opts {
		opt(sho)
	}
	return c.sendHeartbeat(ctx, sho)
}

func (c *Client) sendHeartbeat(ctx context.Context, options *sendHeartbeatOptions) error {
	leaseDuration := c.leaseDuration

	lockItem := options.lockItem
	lockItem.semaphore.Lock()
	defer lockItem.semaphore.Unlock()

	if lockItem.isExpired() || lockItem.ownerName != c.ownerName || lockItem.isReleased {
		c.locks.Delete(lockItem.uniqueIdentifier())
		return &LockNotGrantedError{msg: "cannot send heartbeat because lock is not granted"}
	}

	// Set up condition for UpdateItem. Basically any changes require:
	// 1. I own the lock
	// 2. I know the current version number
	// 3. The lock already exists (UpdateItem API can cause a new item to be created if you do not condition the primary keys with attribute_exists)

	newRvn := c.generateRecordVersionNumber()

	cond := ownershipLockCondition(c.partitionKeyName, lockItem.recordVersionNumber, lockItem.ownerName)
	update := expression.
		Set(leaseDurationAttr, expression.Value(leaseDuration.String())).
		Set(rvnAttr, expression.Value(newRvn))

	if options.deleteData {
		update.Remove(dataAttr)
	} else if len(options.data) > 0 {
		update.Set(dataAttr, expression.Value(options.data))
	}
	updateExpr, _ := expression.NewBuilder().WithCondition(cond).WithUpdate(update).Build()

	updateItemInput := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(c.tableName),
		Key:                       c.getItemKeys(lockItem),
		ConditionExpression:       updateExpr.Condition(),
		UpdateExpression:          updateExpr.Update(),
		ExpressionAttributeNames:  updateExpr.Names(),
		ExpressionAttributeValues: updateExpr.Values(),
	}

	lastUpdateOfLock := time.Now()

	_, err := c.dynamoDB.UpdateItemWithContext(ctx, updateItemInput)
	if err != nil {
		err := parseDynamoDBError(err, "already acquired lock, stopping heartbeats")
		if isLockNotGrantedError(err) {
			c.locks.Delete(lockItem.uniqueIdentifier())
		}
		return err
	}

	lockItem.updateRVN(newRvn, lastUpdateOfLock, leaseDuration)
	return nil
}
