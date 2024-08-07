// Copyright (c) 2016 Uber Technologies, Inc.
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package ctxpropagation

import (
	js "encoding/json"
	"fmt"
	"time"

	"github.com/yarpc/yarpc-go"
	"github.com/yarpc/yarpc-go/crossdock-go"
	"github.com/yarpc/yarpc-go/crossdock/client/params"
	server "github.com/yarpc/yarpc-go/crossdock/server/yarpc"
	"github.com/yarpc/yarpc-go/encoding/json"
	"github.com/yarpc/yarpc-go/internal/baggage"
	"github.com/yarpc/yarpc-go/transport"
	ht "github.com/yarpc/yarpc-go/transport/http"
	tch "github.com/yarpc/yarpc-go/transport/tchannel"

	"github.com/uber/tchannel-go"
	"golang.org/x/net/context"
)

type handler interface {
	Inject(json.Client, server.TransportConfig) // this is hacky but whatever
	Handle(*json.ReqMeta, interface{}) (interface{}, *json.ResMeta, error)
}

func assertBaggageMatches(t crossdock.T, ctx context.Context, want transport.Headers) bool {
	assert := crossdock.Assert(t)
	got := baggage.FromContext(ctx)
	if len(want) == 0 { // nils
		return assert.Len(got, 0, "baggage must be empty: %v", got)
	}

	return assert.Equal(want, got, "baggage must match")
}

// singleHopHandler provides a JSON handler which verifies that it receives the
// specified baggage.
type singleHopHandler struct {
	t           crossdock.T
	wantBaggage transport.Headers
}

func (h *singleHopHandler) Inject(json.Client, server.TransportConfig) {
}

func (h *singleHopHandler) Handle(reqMeta *json.ReqMeta, body interface{}) (interface{}, *json.ResMeta, error) {
	assertBaggageMatches(h.t, reqMeta.Context, h.wantBaggage)
	return map[string]interface{}{}, &json.ResMeta{Headers: reqMeta.Headers}, nil
}

// multiHopHandler provides a JSON handler which verfiies that it receives the
// specified baggage, adds new baggage to the context, and makes a Phone request
// to the Test Subject, requesting a call to a different procedure.
type multiHopHandler struct {
	t crossdock.T

	phoneClient        json.Client
	phoneCallTo        string
	phoneCallTransport server.TransportConfig

	addBaggage  transport.Headers
	wantBaggage transport.Headers
}

func (h *multiHopHandler) Inject(c json.Client, tc server.TransportConfig) {
	h.phoneClient = c
	h.phoneCallTransport = tc
}

func (h *multiHopHandler) Handle(reqMeta *json.ReqMeta, body interface{}) (interface{}, *json.ResMeta, error) {
	if h.phoneClient == nil {
		panic("call Inject() first")
	}

	assertBaggageMatches(h.t, reqMeta.Context, h.wantBaggage)
	for key, value := range h.addBaggage {
		reqMeta.Context = yarpc.WithBaggage(reqMeta.Context, key, value)
	}

	var resp js.RawMessage
	resMeta, err := h.phoneClient.Call(
		&json.ReqMeta{
			Procedure: "phone",
			Headers:   reqMeta.Headers,
			Context:   reqMeta.Context,
		}, &server.PhoneRequest{
			Service:   "ctxclient",
			Procedure: h.phoneCallTo,
			Transport: h.phoneCallTransport,
			Body:      &js.RawMessage{'{', '}'},
		}, &resp)
	return map[string]interface{}{}, resMeta, err
}

// Run verifies that context is propagated across multiple hops.
//
// Behavior parameters:
// - server: Address of the crossdock test subject server.
// - transport: The transport to make requests to the test subject with.
//
// This behavior sets up yet another server in-process which the Phone
// procedure on the test subject is responsible for calling. Usually this
// behavior will receive requests from the Phone procedure on a transport
// different from the one used to call Phone.
func Run(t crossdock.T) {
	checks := crossdock.Checks(t)
	assert := crossdock.Assert(t)
	fatals := crossdock.Fatals(t)

	tests := []struct {
		desc      string
		initCtx   context.Context
		handlers  map[string]handler
		procedure string
	}{
		{
			desc: "no baggage",
			handlers: map[string]handler{
				"hello": &singleHopHandler{
					t:           t,
					wantBaggage: transport.Headers{},
				},
			},
		},
		{
			desc:    "existing baggage",
			initCtx: yarpc.WithBaggage(context.Background(), "token", "42"),
			handlers: map[string]handler{
				"hello": &singleHopHandler{
					t:           t,
					wantBaggage: transport.Headers{"token": "42"},
				},
			},
		},
		{
			desc:      "add baggage",
			procedure: "one",
			handlers: map[string]handler{
				"one": &multiHopHandler{
					t:           t,
					phoneCallTo: "two",
					addBaggage:  transport.Headers{"x": "1"},
					wantBaggage: transport.Headers{},
				},
				"two": &multiHopHandler{
					t:           t,
					phoneCallTo: "three",
					addBaggage:  transport.Headers{"y": "2"},
					wantBaggage: transport.Headers{"x": "1"},
				},
				"three": &singleHopHandler{
					t:           t,
					wantBaggage: transport.Headers{"x": "1", "y": "2"},
				},
			},
		},
		{
			desc:      "add baggage: existing baggage",
			initCtx:   yarpc.WithBaggage(context.Background(), "token", "123"),
			procedure: "one",
			handlers: map[string]handler{
				"one": &multiHopHandler{
					t:           t,
					phoneCallTo: "two",
					addBaggage:  transport.Headers{"hello": "world"},
					wantBaggage: transport.Headers{"token": "123"},
				},
				"two": &singleHopHandler{
					t:           t,
					wantBaggage: transport.Headers{"token": "123", "hello": "world"},
				},
			},
		},
		{
			desc:      "overwrite baggage",
			initCtx:   yarpc.WithBaggage(context.Background(), "x", "1"),
			procedure: "one",
			handlers: map[string]handler{
				"one": &multiHopHandler{
					t:           t,
					phoneCallTo: "two",
					addBaggage:  transport.Headers{"x": "2", "y": "3"},
					wantBaggage: transport.Headers{"x": "1"},
				},
				"two": &multiHopHandler{
					t:           t,
					phoneCallTo: "three",
					addBaggage:  transport.Headers{"y": "4"},
					wantBaggage: transport.Headers{"x": "2", "y": "3"},
				},
				"three": &singleHopHandler{
					t:           t,
					wantBaggage: transport.Headers{"x": "2", "y": "4"},
				},
			},
		},
	}

	for _, tt := range tests {
		func() {
			procedure := tt.procedure
			if procedure == "" {
				if !assert.Len(tt.handlers, 1,
					"%v: invalid test: starting procedure must be provided", tt.desc) {
					return
				}
				for k := range tt.handlers {
					procedure = k
				}
			}

			// We need to build the channel here so that we can have it close in
			// client-only cases.
			//
			// TODO: this shouldn't be an issue once Outbounds support a Close
			// method.
			ch, err := tchannel.NewChannel("ctxclient", nil)
			fatals.NoError(err, "failed to create TChannel")
			defer ch.Close()

			rpc, tconfig := buildRPC(t, ch)
			fatals.NoError(rpc.Start(), "%v: RPC failed to start", tt.desc)
			defer rpc.Stop()

			jsonClient := json.New(rpc.Channel("yarpc-test"))
			for name, handler := range tt.handlers {
				handler.Inject(jsonClient, tconfig)
				json.Register(rpc, json.Procedure(name, handler.Handle))
			}

			ctx := context.Background()
			if tt.initCtx != nil {
				ctx = tt.initCtx
			}
			ctx, _ = context.WithTimeout(ctx, time.Second)

			var resp js.RawMessage
			_, err = jsonClient.Call(
				&json.ReqMeta{Procedure: "phone", Context: ctx},
				&server.PhoneRequest{
					Service:   "ctxclient",
					Procedure: procedure,
					Transport: tconfig,
					Body:      &js.RawMessage{'{', '}'},
				}, &resp)

			checks.NoError(err, "%v: request failed", tt.desc)
		}()
	}
}

func buildRPC(t crossdock.T, ch *tchannel.Channel) (rpc yarpc.RPC, tconfig server.TransportConfig) {
	fatals := crossdock.Fatals(t)

	self := t.Param("ctxclient")
	subject := t.Param("ctxserver")
	fatals.NotEmpty(self, "ctxclient is required")
	fatals.NotEmpty(subject, "ctxserver is required")

	var outbound transport.Outbound
	switch trans := t.Param(params.Transport); trans {
	case "http":
		outbound = ht.NewOutbound(fmt.Sprintf("http://%s:8081", subject))
		tconfig.TChannel = &server.TChannelTransport{Host: self, Port: 8087}
	case "tchannel":
		outbound = tch.NewOutbound(ch, tch.HostPort(fmt.Sprintf("%s:8082", subject)))
		tconfig.HTTP = &server.HTTPTransport{Host: self, Port: 8086}
	default:
		fatals.Fail("", "unknown transport %q", trans)
	}

	rpc = yarpc.New(yarpc.Config{
		Name: "ctxclient",
		Inbounds: []transport.Inbound{
			tch.NewInbound(ch, tch.ListenAddr(":8087")),
			ht.NewInbound(":8086"),
		},
		Outbounds: transport.Outbounds{"yarpc-test": outbound},
	})

	return rpc, tconfig
}
