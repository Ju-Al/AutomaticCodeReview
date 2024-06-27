// Copyright The OpenTelemetry Authors
// TracerConfig is a group of options for a Tracer.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trace

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// tracerConfig is a group of options for a Tracer.
type TracerConfig struct {
	instrumentationVersion string
}

// InstrumentationVersion returns the version of the library providing instrumentation.
func (t *TracerConfig) InstrumentationVersion() string {
	return t.instrumentationVersion
}

// NewTracerConfig applies all the options to a returned TracerConfig.
func NewTracerConfig(options ...TracerOption) *TracerConfig {
	config := new(TracerConfig)
	for _, option := range options {
		option.apply(config)
	}
	return config
}

// TracerOption applies an option to a TracerConfig.
type TracerOption interface {
	apply(*TracerConfig)
}

type tracerOptionFunc func(*TracerConfig)

func (fn tracerOptionFunc) apply(cfg *TracerConfig) {
	fn(cfg)
}

// SpanConfig is a group of options for a Span.
type SpanConfig struct {
	// Attributes describe the associated qualities of a Span.
	Attributes []attribute.KeyValue
	// Timestamp is a time in a Span life-cycle.
	Timestamp time.Time
	// Links are the associations a Span has with other Spans.
	Links []Link
	// NewRoot identifies a Span as the root Span for a new trace. This is
	// commonly used when an existing trace crosses trust boundaries and the
	// remote parent span context should be ignored for security.
	NewRoot bool
	// SpanKind is the role a Span has in a trace.
	SpanKind SpanKind
}

// NewSpanConfig applies all the options to a returned SpanConfig.
// No validation is performed on the returned SpanConfig (e.g. no uniqueness
// checking or bounding of data), it is left to the SDK to perform this
// action.
func NewSpanConfig(options ...SpanOption) *SpanConfig {
	c := new(SpanConfig)
	for _, option := range options {
		option.ApplySpan(c)
	}
	return c
}

// SpanOption applies an option to a SpanConfig.
type SpanOption interface {
	ApplySpan(*SpanConfig)

	// A private method to prevent users implementing the
	// interface and so future additions to it will not
	// violate compatibility.
	private()
}

type EventConfig struct {
	// Attributes describe the associated qualities of a Event.
	Attributes []attribute.KeyValue
	// Timestamp is a time in a Event was recorded.
	Timestamp time.Time
}

// NewEventConfig applies all the EventOptions to a returned SpanConfig. If no
// timestamp option is passed, the returned SpanConfig will have a Timestamp
// set to the call time, otherwise no validation is performed on the returned
// SpanConfig.
func NewEventConfig(options ...EventOption) *SpanConfig {
	c := new(SpanConfig)
	for _, option := range options {
		option.ApplyEvent(c)
	}
	if c.Timestamp.IsZero() {
		c.Timestamp = time.Now()
	}
	return c
}

// EventOption applies span event options to a SpanConfig.
type EventOption interface {
	ApplyEvent(*SpanConfig)

	// A private method to prevent users implementing the
	// interface and so future additions to it will not
	// violate compatibility.
	private()
}

// LifeCycleOption applies span life-cycle options to a SpanConfig. These
// options set values releated to events in a spans life-cycle like starting,
// ending, experiencing an error and other user defined notable events.
type LifeCycleOption interface {
	SpanOption
	EventOption
}

type attributeSpanOption []attribute.KeyValue

func (o attributeSpanOption) ApplySpan(c *SpanConfig)  { o.apply(c) }
func (o attributeSpanOption) ApplyEvent(c *SpanConfig) { o.apply(c) }
func (attributeSpanOption) private()                   {}
func (o attributeSpanOption) apply(c *SpanConfig) {
	c.Attributes = append(c.Attributes, []attribute.KeyValue(o)...)
}

// WithAttributes adds the attributes related to a span life-cycle event.
// These attributes are used to describe the work a Span represents when this
// option is provided to a Span's start or end events. Otherwise, these
// attributes provide additional information about the event being recorded
// (e.g. error, state change, processing progress, system event).
//
// If multiple of these options are passed the attributes of each successive
// option will extend the attributes instead of overwriting. There is no
// guarantee of uniqueness in the resulting attributes.
func WithAttributes(attributes ...attribute.KeyValue) LifeCycleOption {
	return attributeSpanOption(attributes)
}

type timestampSpanOption time.Time

func (o timestampSpanOption) ApplySpan(c *SpanConfig)  { o.apply(c) }
func (o timestampSpanOption) ApplyEvent(c *SpanConfig) { o.apply(c) }
func (timestampSpanOption) private()                   {}
func (o timestampSpanOption) apply(c *SpanConfig)      { c.Timestamp = time.Time(o) }

// WithTimestamp sets the time of a Span life-cycle moment (e.g. started,
// stopped, errored).
func WithTimestamp(t time.Time) LifeCycleOption {
	return timestampSpanOption(t)
}

type linksSpanOption []Link

func (o linksSpanOption) ApplySpan(c *SpanConfig) { c.Links = append(c.Links, []Link(o)...) }
func (linksSpanOption) private()                  {}

// WithLinks adds links to a Span. The links are added to the existing Span
// links, i.e. this does not overwrite.
func WithLinks(links ...Link) SpanOption {
	return linksSpanOption(links)
}

type newRootSpanOption bool

func (o newRootSpanOption) ApplySpan(c *SpanConfig) { c.NewRoot = bool(o) }
func (newRootSpanOption) private()                  {}

// WithNewRoot specifies that the Span should be treated as a root Span. Any
// existing parent span context will be ignored when defining the Span's trace
// identifiers.
func WithNewRoot() SpanOption {
	return newRootSpanOption(true)
}

type spanKindSpanOption SpanKind

func (o spanKindSpanOption) ApplySpan(c *SpanConfig) { c.SpanKind = SpanKind(o) }
func (o spanKindSpanOption) private()                {}

// WithSpanKind sets the SpanKind of a Span.
func WithSpanKind(kind SpanKind) SpanOption {
	return spanKindSpanOption(kind)
}

// WithInstrumentationVersion sets the instrumentation version.
func WithInstrumentationVersion(version string) TracerOption {
	return tracerOptionFunc(func(cfg *TracerConfig) {
		cfg.instrumentationVersion = version
	})
}
