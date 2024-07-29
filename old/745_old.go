type Int64UpDownCounter struct {
	syncInstrument
}

// BoundFloat64UpDownCounter is a bound instrument for Float64UpDownCounter.
//
// It inherits the Unbind function from syncBoundInstrument.
type BoundFloat64UpDownCounter struct {
	syncBoundInstrument
}

// BoundInt64UpDownCounter is a boundInstrument for Int64UpDownCounter.
//
// It inherits the Unbind function from syncBoundInstrument.
type BoundInt64UpDownCounter struct {
	syncBoundInstrument
}

// Bind creates a bound instrument for this counter. The labels are
// associated with values recorded via subsequent calls to Record.
func (c Float64UpDownCounter) Bind(labels ...kv.KeyValue) (h BoundFloat64UpDownCounter) {
	h.syncBoundInstrument = c.bind(labels)
	return
}

// Bind creates a bound instrument for this counter. The labels are
// associated with values recorded via subsequent calls to Record.
func (c Int64UpDownCounter) Bind(labels ...kv.KeyValue) (h BoundInt64UpDownCounter) {
	h.syncBoundInstrument = c.bind(labels)
	return
}

// Measurement creates a Measurement object to use with batch
// recording.
func (c Float64UpDownCounter) Measurement(value float64) Measurement {
	return c.float64Measurement(value)
}

// Measurement creates a Measurement object to use with batch
// recording.
func (c Int64UpDownCounter) Measurement(value int64) Measurement {
	return c.int64Measurement(value)
}

// Add adds the value to the counter's sum. The labels should contain
// the keys and values to be associated with this value.
func (c Float64UpDownCounter) Add(ctx context.Context, value float64, labels ...kv.KeyValue) {
	c.directRecord(ctx, NewFloat64Number(value), labels)
}

// Add adds the value to the counter's sum. The labels should contain
// the keys and values to be associated with this value.
func (c Int64UpDownCounter) Add(ctx context.Context, value int64, labels ...kv.KeyValue) {
	c.directRecord(ctx, NewInt64Number(value), labels)
}

// Add adds the value to the counter's sum using the labels
// previously bound to this counter via Bind()
func (b BoundFloat64UpDownCounter) Add(ctx context.Context, value float64) {
	b.directRecord(ctx, NewFloat64Number(value))
}

// Add adds the value to the counter's sum using the labels
// previously bound to this counter via Bind()
func (b BoundInt64UpDownCounter) Add(ctx context.Context, value int64) {
	b.directRecord(ctx, NewInt64Number(value))
}
