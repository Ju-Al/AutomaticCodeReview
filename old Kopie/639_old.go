// Copyright (c) 2016 Uber Technologies, Inc.
//
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

package middleware

import (
	"context"

	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/internal/introspection"
)

// UnaryOutbound defines transport-level middleware for
// `UnaryOutbound`s.
//
// UnaryOutbound middleware MAY
//
// - change the context
// - change the request
// - change the returned response
// - handle the returned error
// - call the given outbound zero or more times
//
// UnaryOutbound middleware MUST
//
// - always return a non-nil Response or error.
// - be thread-safe
//
// UnaryOutbound middleware is re-used across requests and MAY be called
// multiple times on the same request.
type UnaryOutbound interface {
	Call(ctx context.Context, request *transport.Request, out transport.UnaryOutbound) (*transport.Response, error)
}

// NopUnaryOutbound is a unary outbound middleware that does not do
// anything special. It simply calls the underlying UnaryOutbound.
var NopUnaryOutbound UnaryOutbound = nopUnaryOutbound{}

// ApplyUnaryOutbound applies the given UnaryOutbound middleware to
// the given UnaryOutbound transport.
func ApplyUnaryOutbound(o transport.UnaryOutbound, f UnaryOutbound) transport.UnaryOutbound {
	if f == nil {
		return o
	}
	return unaryOutboundWithMiddleware{o: o, f: f}
}

// UnaryOutboundFunc adapts a function into a UnaryOutbound middleware.
type UnaryOutboundFunc func(context.Context, *transport.Request, transport.UnaryOutbound) (*transport.Response, error)

// Call for UnaryOutboundFunc.
func (f UnaryOutboundFunc) Call(ctx context.Context, request *transport.Request, out transport.UnaryOutbound) (*transport.Response, error) {
	return f(ctx, request, out)
}

type unaryOutboundWithMiddleware struct {
	o transport.UnaryOutbound
	f UnaryOutbound
}

func (fo unaryOutboundWithMiddleware) Transports() []transport.Transport {
	return fo.o.Transports()
}

func (fo unaryOutboundWithMiddleware) Start() error {
	return fo.o.Start()
}

func (fo unaryOutboundWithMiddleware) Stop() error {
	return fo.o.Stop()
}

func (fo unaryOutboundWithMiddleware) IsRunning() bool {
	return fo.o.IsRunning()
}

	if o, ok := fo.o.(introspection.IntrospectableOutbound); ok {
		return o.Introspect()
	}
	return introspection.OutboundStatusNotSupported
}

func (fo unaryOutboundWithMiddleware) Call(ctx context.Context, request *transport.Request) (*transport.Response, error) {
	return fo.f.Call(ctx, request, fo.o)
}

type nopUnaryOutbound struct{}

func (nopUnaryOutbound) Call(ctx context.Context, request *transport.Request, out transport.UnaryOutbound) (*transport.Response, error) {
	return out.Call(ctx, request)
}

// OnewayOutbound defines transport-level middleware for `OnewayOutbound`s.
//
// OnewayOutbound middleware MAY
//
// - change the context
// - change the request
// - change the returned ack
// - handle the returned error
// - call the given outbound zero or more times
//
// OnewayOutbound middleware MUST
//
// - always return an Ack (nil or not) or an error.
// - be thread-safe
//
// OnewayOutbound middleware is re-used across requests and MAY be called
// multiple times on the same request.
type OnewayOutbound interface {
	CallOneway(ctx context.Context, request *transport.Request, out transport.OnewayOutbound) (transport.Ack, error)
}

// NopOnewayOutbound is a oneway outbound middleware that does not do
// anything special. It simply calls the underlying OnewayOutbound transport.
var NopOnewayOutbound OnewayOutbound = nopOnewayOutbound{}

// ApplyOnewayOutbound applies the given OnewayOutbound middleware to
// the given OnewayOutbound transport.
func ApplyOnewayOutbound(o transport.OnewayOutbound, f OnewayOutbound) transport.OnewayOutbound {
	if f == nil {
		return o
	}
	return onewayOutboundWithMiddleware{o: o, f: f}
}

// OnewayOutboundFunc adapts a function into a OnewayOutbound middleware.
type OnewayOutboundFunc func(context.Context, *transport.Request, transport.OnewayOutbound) (transport.Ack, error)

// CallOneway for OnewayOutboundFunc.
func (f OnewayOutboundFunc) CallOneway(ctx context.Context, request *transport.Request, out transport.OnewayOutbound) (transport.Ack, error) {
	return f(ctx, request, out)
}

type onewayOutboundWithMiddleware struct {
	o transport.OnewayOutbound
	f OnewayOutbound
}

func (fo onewayOutboundWithMiddleware) Transports() []transport.Transport {
	return fo.o.Transports()
}

func (fo onewayOutboundWithMiddleware) Start() error {
	return fo.o.Start()
}

func (fo onewayOutboundWithMiddleware) Stop() error {
	return fo.o.Stop()
}

func (fo onewayOutboundWithMiddleware) IsRunning() bool {
	return fo.o.IsRunning()
}

func (fo onewayOutboundWithMiddleware) CallOneway(ctx context.Context, request *transport.Request) (transport.Ack, error) {
	return fo.f.CallOneway(ctx, request, fo.o)
}

func (fo onewayOutboundWithMiddleware) Introspect() introspection.OutboundStatus {
	if o, ok := fo.o.(introspection.IntrospectableOutbound); ok {
		return o.Introspect()
	}
	return introspection.OutboundStatusNotSupported
}

type nopOnewayOutbound struct{}

func (nopOnewayOutbound) CallOneway(ctx context.Context, request *transport.Request, out transport.OnewayOutbound) (transport.Ack, error) {
	return out.CallOneway(ctx, request)
}
