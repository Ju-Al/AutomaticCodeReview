// Copyright 2020 The Swarm Authors. All rights reserved.
	"sync"
	"github.com/ethersphere/bee/pkg/pushsync"
	"github.com/ethersphere/bee/pkg/pushsync/pb"
	"github.com/ethersphere/bee/pkg/topology"
	"github.com/ethersphere/bee/pkg/topology/mock"

// TestSendToClosest tests that a chunk that is uploaded to localstore is sent to the appropriate peer.
func TestSendToClosest(t *testing.T) {
	logger := logging.New(ioutil.Discard, 0)

	// Create a mock connectivity between the peers
	mockTopology := mock.NewTopologyDriver(mock.WithClosestPeer(closestPeer))

	storer, err := localstore.New("", pivotNode.Bytes(), nil, logger)
	if err != nil {
		t.Fatal(err)
	}
	// setup the stream recorder to record stream data
			return func(context.Context, p2p.Peer, p2p.Stream) error {
				// dont call any handlers
				return nil
			}
	// instantiate a pushsync instance
	ps := pushsync.New(pushsync.Options{
		Streamer:      recorder,
		Logger:        logger,
		ClosestPeerer: mockTopology,
		Storer:        storer,
	})
	defer ps.Close()
	recorder.SetProtocols(ps.Protocol())
	_, err = storer.Put(context.Background(), storage.ModePutUpload, swarm.NewChunk(chunkAddress, chunkData))
	records := recorder.WaitRecords(t, closestPeer, pushsync.ProtocolName, pushsync.ProtocolVersion, pushsync.StreamName, 1, 5)
	messages, err := protobuf.ReadMessages(
		bytes.NewReader(records[0].In()),
		func() protobuf.Message { return new(pb.Delivery) },
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) > 1 {
		t.Fatal("too many messages")
	}
	delivery := messages[0].(*pb.Delivery)
	chunk := swarm.NewChunk(swarm.NewAddress(delivery.Address), delivery.Data)
	if !bytes.Equal(chunk.Address().Bytes(), chunkAddress.Bytes()) {
		t.Fatalf("chunk address mismatch")
	}
	if !bytes.Equal(chunk.Data(), chunkData) {
		t.Fatalf("chunk data mismatch")
	}
// TestForwardChunk tests that when a closer node exists within the topology, we forward a received
// chunk to it.
func TestForwardChunk(t *testing.T) {
	logger := logging.New(ioutil.Discard, 0)

	// create a pivot node and a closest mocked closer node address
	pivotNode := swarm.MustParseHexAddress("0000000000000000000000000000000000000000000000000000000000000000")   // pivot is 0000
	closestPeer := swarm.MustParseHexAddress("6000000000000000000000000000000000000000000000000000000000000000") // binary 0110

	// Create a mock connectivity driver
	mockTopology := mock.NewTopologyDriver(mock.WithClosestPeer(closestPeer))
	storer, err := localstore.New("", pivotNode.Bytes(), nil, logger)
	if err != nil {
		t.Fatal(err)
	}
	targetCalled := false
	var mtx sync.Mutex
	// setup the stream recorder to record stream data
	recorder := streamtest.New(

			// this is a custom middleware that is needed because of the design of
			// the recorder. since we want to test only one unit, but the first message
			// is supposedly coming from another node, we don't want to execute the handler
			// when the peer address is the peer of `closestPeer`, since this will create an
			// unnecessary entry in the recorder
			return func(ctx context.Context, p p2p.Peer, s p2p.Stream) error {
				if p.Address.Equal(closestPeer) {
					mtx.Lock()
					defer mtx.Unlock()
					if targetCalled {
						t.Fatal("target called more than once")
					}
					targetCalled = true
					return nil
				}
				return f(ctx, p, s)
			}
	ps := pushsync.New(pushsync.Options{
		Streamer:      recorder,
		Logger:        logger,
		ClosestPeerer: mockTopology,
		Storer:        storer,
	})
	defer ps.Close()
	recorder.SetProtocols(ps.Protocol())
	stream, err := recorder.NewStream(context.Background(), pivotNode, nil, pushsync.ProtocolName, pushsync.ProtocolVersion, pushsync.StreamName)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	w := protobuf.NewWriter(stream)
	// this triggers the handler of the pivot with a delivery stream
	err = w.WriteMsg(&pb.Delivery{
		Address: chunkAddress.Bytes(),
		Data:    chunkData,
	})
	_ = recorder.WaitRecords(t, closestPeer, pushsync.ProtocolName, pushsync.ProtocolVersion, pushsync.StreamName, 1, 5)
	mtx.Lock()
	defer mtx.Unlock()
	if !targetCalled {
		t.Fatal("target not called")
	}
}
// TestNoForwardChunk tests that the closest node to a chunk doesn't forward it to other nodes.
func TestNoForwardChunk(t *testing.T) {
	logger := logging.New(ioutil.Discard, 0)
	// chunk data to upload
	chunkAddress := swarm.MustParseHexAddress("7000000000000000000000000000000000000000000000000000000000000000") // binary 0111
	chunkData := []byte("1234")
	// create a pivot node and a cluster of nodes
	pivotNode := swarm.MustParseHexAddress("6000000000000000000000000000000000000000000000000000000000000000") // pivot is 0110
	// Create a mock connectivity
	mockTopology := mock.NewTopologyDriver(mock.WithClosestPeerErr(topology.ErrWantSelf))
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pushsync_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	"github.com/ethersphere/bee/pkg/localstore"
	"github.com/ethersphere/bee/pkg/logging"
	"github.com/ethersphere/bee/pkg/p2p"
	"github.com/ethersphere/bee/pkg/p2p/protobuf"
	"github.com/ethersphere/bee/pkg/p2p/streamtest"
	"github.com/ethersphere/bee/pkg/pushsync"
	"github.com/ethersphere/bee/pkg/pushsync/pb"
	"github.com/ethersphere/bee/pkg/storage"
	"github.com/ethersphere/bee/pkg/swarm"
	"github.com/ethersphere/bee/pkg/topology"
	"github.com/ethersphere/bee/pkg/topology/mock"
)

// TestSendChunkAndGetReceipt inserts a chunk as uploaded chunk in db. This triggers sending a chunk to the closest node
// and expects a receipt. The message are intercepted in the outgoing stream to check for correctness.
func TestSendChunkAndReceiveReceipt(t *testing.T) {
	// chunk data to upload
	chunkAddress := swarm.MustParseHexAddress("7000000000000000000000000000000000000000000000000000000000000000")
	chunkData := []byte("1234")
	chunk := swarm.NewChunk(chunkAddress, chunkData)

	// create a pivot node and a mocked closest node
	pivotNode := swarm.MustParseHexAddress("0000000000000000000000000000000000000000000000000000000000000000")   // base is 0000
	closestPeer := swarm.MustParseHexAddress("6000000000000000000000000000000000000000000000000000000000000000") // binary 0110 -> po 1

	// peer is the node responding to the chunk receipt message
	// mock should return ErrWantSelf since there's no one to forward to
	psPeer, storerPeer := createPushSyncNode(t, closestPeer, nil, mock.WithClosestPeerErr(topology.ErrWantSelf))
	defer storerPeer.Close()
	defer psPeer.Close()

	recorder := streamtest.New(
		streamtest.WithProtocols(psPeer.Protocol()),
		streamtest.WithMiddlewares(func(f p2p.HandlerFunc) p2p.HandlerFunc {
			return f
		}),
	)

	// pivot node needs the streamer since the chunk is intercepted by
	// the chunk worker, then gets sent by opening a new stream
	psPivot, storerPivot := createPushSyncNode(t, pivotNode, recorder, mock.WithClosestPeer(closestPeer))

	// upload the chunk to the pivot node
	_, err := storerPivot.Put(context.Background(), storage.ModePutUpload, chunk)
	if err != nil {
		t.Fatal(err)
	}

	// this intercepts the outgoing delivery message
	waitOnRecordAndTest(t, closestPeer, recorder, chunkAddress, chunkData)

	// this intercepts the incoming receipt message
	waitOnRecordAndTest(t, closestPeer, recorder, chunkAddress, nil)

	// Close the pushsync and then the DB
	psPivot.Close()
	storerPivot.Close()
}

// TestHandler expect a chunk from a node on a stream. It then stores the chunk in the local store and
// sends back a receipt. This is tested by intercepting the incoming stream for proper messages.
// It also sends the chunk to the closest peerand receives a receipt.
//
// Chunk moves from   TriggerPeer -> PivotPeer -> ClosestPeer
//
func TestHandler(t *testing.T) {
	// chunk data to upload
	chunkAddress := swarm.MustParseHexAddress("7000000000000000000000000000000000000000000000000000000000000000")
	chunkData := []byte("1234")
	chunk := swarm.NewChunk(chunkAddress, chunkData)

	// create a pivot node and a mocked closest node
	pivotPeer := swarm.MustParseHexAddress("0000000000000000000000000000000000000000000000000000000000000000")
	triggerPeer := swarm.MustParseHexAddress("6000000000000000000000000000000000000000000000000000000000000000")
	closestPeer := swarm.MustParseHexAddress("f000000000000000000000000000000000000000000000000000000000000000")

	// Create the closest peer
	psClosestPeer, closestStorerPeerDB := createPushSyncNode(t, closestPeer, nil, mock.WithClosestPeerErr(topology.ErrWantSelf))

	closestRecorder := streamtest.New(
		streamtest.WithProtocols(psClosestPeer.Protocol()),
		streamtest.WithMiddlewares(func(f p2p.HandlerFunc) p2p.HandlerFunc {
			return f
		}),
	)

	// creating the pivot peer
	psPivot, storerPivotDB := createPushSyncNode(t, pivotPeer, closestRecorder, mock.WithClosestPeer(closestPeer))

	pivotRecorder := streamtest.New(
		streamtest.WithProtocols(psPivot.Protocol()),
		streamtest.WithMiddlewares(func(f p2p.HandlerFunc) p2p.HandlerFunc {
			return f
		}),
	)

	// Creating the trigger peer
	psTriggerPeer, triggerStorerDB := createPushSyncNode(t, triggerPeer, pivotRecorder, mock.WithClosestPeer(pivotPeer))

	// upload the chunk to the trigger node DB which triggers the chain reaction of sending this chunk
	// from trigger peer to pivot peer to closest peer.
	_, err := triggerStorerDB.Put(context.Background(), storage.ModePutUpload, chunk)
	if err != nil {
		t.Fatal(err)
	}

	// In pivot peer,  intercept the incoming delivery chunk from the trigger peer and check for correctness
	waitOnRecordAndTest(t, pivotPeer, pivotRecorder, chunkAddress, chunkData)

	// Pivot peer will forward the chunk to its closest peer. Intercept the incoming stream from pivot node and check
	// for the correctness of the chunk
	waitOnRecordAndTest(t, closestPeer, closestRecorder, chunkAddress, chunkData)

	// Similarly intercept the same incoming stream to see if the closest peer is sending a proper receipt
	waitOnRecordAndTest(t, closestPeer, closestRecorder, chunkAddress, nil)

	// In the received stream, check if a receipt is sent from pivot peer and check for its correctness.
	waitOnRecordAndTest(t, pivotPeer, pivotRecorder, chunkAddress, nil)

	// closing the ps so that the DB is closed after that to avoid data race
	psClosestPeer.Close()
	psPivot.Close()
	psTriggerPeer.Close()

	// closing all DBs
	closestStorerPeerDB.Close()
	storerPivotDB.Close()
	triggerStorerDB.Close()
}

func createPushSyncNode(t *testing.T, addr swarm.Address, recorder *streamtest.Recorder, mockOpts ...mock.Option) (*pushsync.PushSync, *localstore.DB) {
	logger := logging.New(ioutil.Discard, 0)

	storer, err := localstore.New("", addr.Bytes(), nil, logger)
	if err != nil {
		t.Fatal(err)
	}

	mockTopology := mock.NewTopologyDriver(mockOpts...)

	ps := pushsync.New(pushsync.Options{
		Streamer:      recorder,
		Storer:        storer,
		ClosestPeerer: mockTopology,
		Logger:        logger,
	})

	return ps, storer
}

func waitOnRecordAndTest(t *testing.T, peer swarm.Address, recorder *streamtest.Recorder, add swarm.Address, data []byte) {
	t.Helper()
	records := recorder.WaitRecords(t, peer, pushsync.ProtocolName, pushsync.ProtocolVersion, pushsync.StreamName, 1, 5)

	if data != nil {
		messages, err := protobuf.ReadMessages(
			bytes.NewReader(records[0].In()),
			func() protobuf.Message { return new(pb.Delivery) },
		)
		if err != nil {
			t.Fatal(err)
		}
		if messages == nil {
			t.Fatal("nil rcvd. for message")
		}
		if len(messages) > 1 {
			t.Fatal("too many messages")
		}
		delivery := messages[0].(*pb.Delivery)

		if !bytes.Equal(delivery.Address, add.Bytes()) {
			t.Fatalf("chunk address mismatch")
		}

		if !bytes.Equal(delivery.Data, data) {
			t.Fatalf("chunk data mismatch")
		}
	} else {
		messages, err := protobuf.ReadMessages(
			bytes.NewReader(records[0].In()),
			func() protobuf.Message { return new(pb.Receipt) },
		)
		if err != nil {
			t.Fatal(err)
		}
		if messages == nil {
			t.Fatal("nil rcvd. for message")
		}
		if len(messages) > 1 {
			t.Fatal("too many messages")
		}
		receipt := messages[0].(*pb.Receipt)
		receiptAddress := swarm.NewAddress(receipt.Address)

		if !receiptAddress.Equal(add) {
			t.Fatalf("receipt address mismatch")
		}
	}

}
