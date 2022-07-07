/*
Copyright 2015 github.com/ucirello

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
	"errors"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// Lock item properly speaking.
type Lock struct {
	semaphore sync.Mutex

	client       *Client
	partitionKey string

	data                []byte
	ownerName           string
	deleteLockOnRelease bool
	isReleased          bool
	sessionMonitor      *sessionMonitor

	lookupTime           time.Time
	recordVersionNumber  string
	leaseDuration        time.Duration
	additionalAttributes map[string]*dynamodb.AttributeValue
}

// Data returns the content of the lock, if any is available.
func (l *Lock) Data() []byte {
	if l == nil {
		return nil
	}
	return l.data
}

// Close releases the lock.
func (l *Lock) Close() error {
	if l != nil && l.client != nil {
		if l.IsExpired() {
			return ErrLockAlreadyReleased
		}
		_, err := l.client.ReleaseLock(l)
		return err
	}
	return ErrCannotReleaseNullLock
}

func (l *Lock) uniqueIdentifier() string {
	return l.partitionKey
}

// IsExpired returns if the lock is expired, released, or neither.
func (l *Lock) IsExpired() bool {
	if l == nil {
		return true
	}
	l.semaphore.Lock()
	defer l.semaphore.Unlock()
	return l.isExpired()
}

func (l *Lock) isExpired() bool {
	if l == nil {
		return true
	}

	if l.isReleased {
		return true
	}
	return time.Since(l.lookupTime) > l.leaseDuration
}

func (l *Lock) updateRVN(rvn string, lastUpdate time.Time, leaseDuration time.Duration) {
	l.recordVersionNumber = rvn
	l.lookupTime = lastUpdate
	l.leaseDuration = leaseDuration
}

// OwnerName returns the lock's owner.
func (l *Lock) OwnerName() string {
	if l == nil {
		return ""
	}
	return l.ownerName
}

// AdditionalAttributes returns the lock's additional data stored during
// acquisition.
func (l *Lock) AdditionalAttributes() map[string]*dynamodb.AttributeValue {
	addAttr := make(map[string]*dynamodb.AttributeValue)
	if l != nil {
		for k, v := range l.additionalAttributes {
			addAttr[k] = v
		}
	}
	return addAttr
}

// IsAlmostExpired returns whether or not the lock is entering the "danger
// zone" time period.
//
// It returns if the lock has been released or the lock's lease has entered the
// "danger zone". It returns false if the lock has not been released and the
// lock has not yet entered the "danger zone"
func (l *Lock) IsAlmostExpired() (bool, error) {
	t, err := l.timeUntilDangerZoneEntered()
	if err != nil {
		return false, err
	}
	return t <= 0, nil
}

// Errors related to session manager life-cycle.
var (
	ErrSessionMonitorNotSet  = errors.New("session monitor is not set")
	ErrLockAlreadyReleased   = errors.New("lock is already released")
	ErrCannotReleaseNullLock = errors.New("cannot release null lock item")
	ErrOwnerMismatched       = errors.New("lock owner mismatched")
)

func (l *Lock) timeUntilDangerZoneEntered() (time.Duration, error) {
	if l == nil {
		return 0, ErrLockAlreadyReleased
	}
	if l.sessionMonitor == nil {
		return 0, ErrSessionMonitorNotSet
	}
	if l.IsExpired() {
		return 0, ErrLockAlreadyReleased
	}
	return l.sessionMonitor.timeUntilLeaseEntersDangerZone(l.lookupTime), nil
}
