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

package natspubsub_test

import (
	"context"
	"log"

	"github.com/nats-io/go-nats"
	"gocloud.dev/pubsub"
	"gocloud.dev/pubsub/natspubsub"
)

func ExampleOpenTopic() {
	ctx := context.Background()

	// Create a connection to NATS.
	natsConn, err := nats.Connect("nats://demo.nats.io")
	// To use NGS with credentials (https://synadia.com/ngs):
	// natsConn, err := nats.Connect("connect.ngs/global", nats.UserCredentials("path_to_creds_file")
	if err != nil {
		log.Fatal(err)
	}
	defer natsConn.Close()

	// Open a pubsub.Topic that publishes message with subject "go-cloud.example.send".
	topic, err := natspubsub.OpenTopic(natsConn, "go-cloud.example.send", nil)
	if err != nil {
		// Handle error....
	}

	err = topic.Send(ctx, &pubsub.Message{Body: []byte("example message")})
}

func ExampleOpenSubscription() {
	ctx := context.Background()

	// Create a connection to NATS.
	natsConn, err := nats.Connect("nats://demo.nats.io")
	// To use NGS with credentials (https://synadia.com/ngs):
	// natsConn, err := nats.Connect("connect.ngs/global", nats.UserCredentials("path_to_creds_file")
	if err != nil {
		log.Fatal(err)
	}
	defer natsConn.Close()

	ackFunc := func() {
		// This function will be called when the application calls "Ack" on a
		// received message.
		// Since Ack is a meaningless no-op for NATS, you can provide an empty
		// function to do nothing, or panic/log a warning if your application
		// is built for at-most-once semantics and should never call Ack.
		// See https://godoc.org/gocloud.dev/pubsub#hdr-At_most_once_and_At_least_once_Delivery
		// for more details.
	}
	// Open a pubsub.Subscription that receives messages with subject "go-cloud.example.receive".
	sub, err := natspubsub.OpenSubscription(natsConn, "go-cloud.example.receive", ackFunc, nil)
	if err != nil {
		// Handle error....
	}

	// Now we can use sub to receive messages.
	msg, err := sub.Receive(ctx)
	if err != nil {
		// Handle the error....
	}
	// Handle the message....

	// Ack will call ackFunc above. If you're only going to use at-most-once
	// providers, you can omit it.
	msg.Ack()
}

func Example_openFromURL() {
	ctx := context.Background()

	// OpenTopic creates a *pubsub.Topic from a URL.
	// This URL will Dial the NATS server at the URL in the environment
	// variable NATS_SERVER_URL and send messages with subject "mytopic".
	t, err := pubsub.OpenTopic(ctx, "nats://mytopic")

	// Similarly, OpenSubscription creates a *pubsub.Subscription from a URL.
	// This URL will use the same connection and receive messages with subject
	// "mytopic".
	// Note that by default, s.Ack will result in a panic, as Ack is a meaningless
	// no-op for NATS. You can disable the panic using "?ackfunc=log" or
	// "?ackfunc=noop".
	sub, err := pubsub.OpenSubscription(ctx, "nats://mytopic")
	_, _, _ = topic, sub, err
}
