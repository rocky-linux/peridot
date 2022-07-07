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

// Package dynamolock provides a simple utility for using DynamoDB's consistent
// read/write feature to use it for managing distributed locks.
//
// In order to use this package, the client must create a table in DynamoDB,
// although the client provides a convenience method for creating that table
// (CreateTable).
//
// Basic usage:
//
//	import (
//		"log"
//
//		"cirello.io/dynamolock"
//		"github.com/aws/aws-sdk-go/aws"
//		"github.com/aws/aws-sdk-go/aws/session"
//		"github.com/aws/aws-sdk-go/service/dynamodb"
//	)
//
//	// ---
//
//	svc := dynamodb.New(session.Must(session.NewSession(&aws.Config{
//		Region: aws.String("us-west-2"),
//	})))
//	c, err := dynamolock.New(svc,
//		"locks",
//		dynamolock.WithLeaseDuration(3*time.Second),
//		dynamolock.WithHeartbeatPeriod(1*time.Second),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer c.Close()
//
//	log.Println("ensuring table exists")
//	c.CreateTable("locks",
//		dynamolock.WithProvisionedThroughput(&dynamodb.ProvisionedThroughput{
//			ReadCapacityUnits:  aws.Int64(5),
//			WriteCapacityUnits: aws.Int64(5),
//		}),
//		dynamolock.WithCustomPartitionKeyName("key"),
//	)
//
//	data := []byte("some content a")
//	lockedItem, err := c.AcquireLock("spock",
//		dynamolock.WithData(data),
//		dynamolock.ReplaceData(),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	log.Println("lock content:", string(lockedItem.Data()))
//	if got := string(lockedItem.Data()); string(data) != got {
//		log.Println("losing information inside lock storage, wanted:", string(data), " got:", got)
//	}
//
//	log.Println("cleaning lock")
//	success, err := c.ReleaseLock(lockedItem)
//	if !success {
//		log.Fatal("lost lock before release")
//	}
//	if err != nil {
//		log.Fatal("error releasing lock:", err)
//	}
//	log.Println("done")
//
// This package is covered by this SLA:
// https://github.com/cirello-io/public/blob/master/SLA.md
//
package dynamolock // import "cirello.io/dynamolock"
