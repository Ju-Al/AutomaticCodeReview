/*
Copyright 2020 Google LLC

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

package ingress

import (
	"context"
	"errors"
	"fmt"
	nethttp "net/http"
	"strings"
	"time"

	_ "knative.dev/pkg/metrics/testing"

	cev2 "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/transformer"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/google/knative-gcp/pkg/metrics"
	"github.com/google/wire"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/kncloudevents"
	"knative.dev/eventing/pkg/logging"
)

const (
	// TODO(liu-cong) configurable timeout
	decoupleSinkTimeout = 30 * time.Second

	// EventArrivalTime is used to access the metadata stored on a
	// CloudEvent to measure the time difference between when an event is
	// received on a broker and before it is dispatched to the trigger function.
	// The format is an RFC3339 time in string format. For example: 2019-08-26T23:38:17.834384404Z.
	EventArrivalTime = "knativearrivaltime"
)

// HandlerSet provides a handler with a real HTTPMessageReceiver and pubsub MultiTopicDecoupleSink.
var HandlerSet wire.ProviderSet = wire.NewSet(
	NewHandler,
	NewHTTPMessageReceiver,
	wire.Bind(new(HttpMessageReceiver), new(*kncloudevents.HttpMessageReceiver)),
	NewMultiTopicDecoupleSink,
	wire.Bind(new(DecoupleSink), new(*multiTopicDecoupleSink)),
	NewPubsubClient,
	NewPubsubDecoupleClient,
	metrics.NewIngressReporter,
)

// DecoupleSink is an interface to send events to a decoupling sink (e.g., pubsub).
type DecoupleSink interface {
	// Send sends the event from a broker to the corresponding decoupling sink.
	Send(ctx context.Context, ns, broker string, event cev2.Event) protocol.Result
}

// HttpMessageReceiver is an interface to listen on http requests.
type HttpMessageReceiver interface {
	StartListen(ctx context.Context, handler nethttp.Handler) error
}

// handler receives events and persists them to storage (pubsub).
type Handler struct {
	// httpReceiver is an HTTP server to receive events.
	httpReceiver HttpMessageReceiver
	// decouple is the client to send events to a decouple sink.
	decouple DecoupleSink
	logger   *zap.Logger
	reporter *metrics.IngressReporter
}

// NewHandler creates a new ingress handler.
func NewHandler(ctx context.Context, httpReceiver HttpMessageReceiver, decouple DecoupleSink, reporter *metrics.IngressReporter) *Handler {
	return &Handler{
		httpReceiver: httpReceiver,
		decouple:     decouple,
		reporter:     reporter,
		logger:       logging.FromContext(ctx),
	}
}

// Start blocks to receive events over HTTP.
func (h *Handler) Start(ctx context.Context) error {
	return h.httpReceiver.StartListen(ctx, h)
}

// ServeHTTP implements net/http Handler interface method.
// 1. Performs basic validation of the request.
// 2. Parse request URL to get namespace and broker.
// 3. Convert request to event.
// 4. Send event to decouple sink.
func (h *Handler) ServeHTTP(response nethttp.ResponseWriter, request *nethttp.Request) {
	ctx := request.Context()
	h.logger.Debug("Serving http", zap.Any("headers", request.Header))
	startTime := time.Now()
	if request.Method != nethttp.MethodPost {
		response.WriteHeader(nethttp.StatusMethodNotAllowed)
		return
	}

	// Path should be in the form of "/<ns>/<broker>".
	pieces := strings.Split(request.URL.Path, "/")
	if len(pieces) != 3 {
		msg := fmt.Sprintf("Malformed request path. want: '/<ns>/<broker>'; got: %v..", request.URL.Path)
		h.logger.Info(msg)
		nethttp.Error(response, msg, nethttp.StatusNotFound)
		return
	}
	broker := types.NamespacedName{
		Namespace: pieces[1],
		Name:      pieces[2],
	}

	event, err := h.toEvent(request)
	if err != nil {
		nethttp.Error(response, err.Error(), nethttp.StatusBadRequest)
		return
	}

	event.SetExtension(EventArrivalTime, cev2.Timestamp{Time: time.Now()})

	ctx, span := trace.StartSpan(ctx, kntracing.BrokerMessagingDestination(broker))
	defer span.End()
	if span.IsRecordingEvents() {
		span.AddAttributes(
			kntracing.MessagingSystemAttribute,
			tracing.PubSubProtocolAttribute,
			kntracing.BrokerMessagingDestinationAttribute(broker),
			kntracing.MessagingMessageIDAttribute(event.ID()),
		)
	}

	// Optimistically set status code to StatusAccepted. It will be updated if there is an error.
	// According to the data plane spec (https://github.com/knative/eventing/blob/master/docs/spec/data-plane.md), a
	// non-callable SINK (which broker is) MUST respond with 202 Accepted if the request is accepted.
	statusCode := nethttp.StatusAccepted
	ctx, cancel := context.WithTimeout(ctx, decoupleSinkTimeout)
	defer cancel()
	defer func() { h.reportMetrics(request.Context(), broker, event, statusCode, startTime) }()
	if res := h.decouple.Send(ctx, broker.Namespace, broker.Name, *event); !cev2.IsACK(res) {
		msg := fmt.Sprintf("Error publishing to PubSub for broker %s. event: %+v, err: %v.", broker, event, res)
		h.logger.Error(msg)
		statusCode = nethttp.StatusInternalServerError
		if errors.Is(res, ErrNotFound) {
			statusCode = nethttp.StatusNotFound
		}
		nethttp.Error(response, msg, statusCode)
		return
	}

	response.WriteHeader(statusCode)
}

// toEvent converts an http request to an event.
func (h *Handler) toEvent(request *nethttp.Request) (*cev2.Event, error) {
	message := http.NewMessageFromHttpRequest(request)
	defer func() {
		if err := message.Finish(nil); err != nil {
			h.logger.Error("Failed to close message", zap.Any("message", message), zap.Error(err))
		}
	}()
	// If encoding is unknown, the message is not an event.
	if message.ReadEncoding() == binding.EncodingUnknown {
		msg := fmt.Sprintf("Encoding is unknown. Not a cloud event? request: %+v", request)
		h.logger.Debug(msg)
		return nil, errors.New(msg)
	}
	event, err := binding.ToEvent(request.Context(), message, transformer.AddTimeNow)
	if err != nil {
		msg := fmt.Sprintf("Failed to convert request to event: %v", err)
		h.logger.Error(msg)
		return nil, errors.New(msg)
	}
	return event, nil
}

func (h *Handler) reportMetrics(ctx context.Context, broker types.NamespacedName, event *cev2.Event, statusCode int, start time.Time) {
	args := metrics.IngressReportArgs{
		Namespace:    broker.Namespace,
		Broker:       broker.Name,
		EventType:    event.Type(),
		ResponseCode: statusCode,
	}
	if err := h.reporter.ReportEventDispatchTime(ctx, args, time.Since(start)); err != nil {
		h.logger.Warn("Failed to record metrics.", zap.Any("namespace", broker.Namespace), zap.Any("broker", broker.Name), zap.Error(err))
	}
}
