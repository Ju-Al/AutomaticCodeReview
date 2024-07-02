// Copyright 2019 The Go Cloud Development Kit Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mongodocstore

// To run these tests against a real MongoDB server, first run:
//     docker run -d --name my-mongo  -p 27017:27017 mongo:4
// Then wait a few seconds for the server to be ready.

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gocloud.dev/internal/docstore/driver"
	"gocloud.dev/internal/docstore/drivertest"
)

const (
	serverURI      = "mongodb://localhost"
	dbName         = "docstore-test"
	collectionName = "docstore-test"
)

type harness struct {
	db *mongo.Database
}

// Return a new mongoDB client that is connected to the server URI.
func newClient(ctx context.Context, uri string) (*mongo.Client, error) {
	opts := options.Client().ApplyURI(uri)
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	client, err := mongo.NewClient(opts)
	if err != nil {
		return nil, err
	}
	if err := client.Connect(ctx); err != nil {
		return nil, err
	}
	// Connect doesn't seem to actually make a connection, so do an RPC.
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	return client, nil
}

func (h *harness) MakeCollection(ctx context.Context) (driver.Collection, error) {
	coll := newCollection(h.db.Collection(collectionName))
	// It seems that the client doesn't actually connect until the first RPC, which will
	// be this one. So time out quickly if there's a problem.
	tctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := coll.coll.Drop(tctx); err != nil {
		return nil, err
	}
	return coll, nil
}

func (h *harness) Close() {}

}

func (codecTester) UnsupportedTypes() []drivertest.UnsupportedType {
	return []drivertest.UnsupportedType{
		drivertest.Complex, drivertest.NanosecondTimes}
}

func (codecTester) DocstoreEncode(x interface{}) (interface{}, error) {
	doc, err := driver.NewDocument(x)
	if err != nil {
		return nil, err
	}
	m, err := encodeDoc(doc)
	if err != nil {
		return nil, err
	}
	return bson.Marshal(m)
}

func (codecTester) DocstoreDecode(value, dest interface{}) error {
	doc, err := driver.NewDocument(dest)
	if err != nil {
		return err
	}
	var m map[string]interface{}
	if err := bson.Unmarshal(value.([]byte), &m); err != nil {
		return err
	}
	return decodeDoc(m, doc)
}

func (codecTester) NativeEncode(x interface{}) (interface{}, error) {
	return bson.Marshal(x)
}

func (codecTester) NativeDecode(value, dest interface{}) error {
	return bson.Unmarshal(value.([]byte), dest)
}

func TestConformance(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	client, err := newClient(ctx, serverURI)
	if err == context.DeadlineExceeded {
		t.Skip("could not connect to local mongoDB server (connection timed out)")
	}
	if err != nil {
		t.Fatalf("connecting to %s: %v", serverURI, err)
	}
	defer func() { client.Disconnect(context.Background()) }()

	newHarness := func(context.Context, *testing.T) (drivertest.Harness, error) {
		return &harness{client.Database(dbName)}, nil
	}
	drivertest.RunConformanceTests(t, newHarness, codecTester{})
}
