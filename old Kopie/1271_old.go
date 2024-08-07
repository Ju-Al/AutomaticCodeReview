// Copyright The OpenTelemetry Authors
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

package metric

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	api "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/global"
	internal "go.opentelemetry.io/otel/internal/metric"
	"go.opentelemetry.io/otel/label"
	export "go.opentelemetry.io/otel/sdk/export/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregator"
	"go.opentelemetry.io/otel/sdk/resource"
)

type (
	// Accumulator implements the OpenTelemetry Meter API.  The
	// Accumulator is bound to a single export.Processor in
	// `NewAccumulator()`.
	//
	// The Accumulator supports a Collect() API to gather and export
	// current data.  Collect() should be arranged according to
	// the processor model.  Push-based processors will setup a
	// timer to call Collect() periodically.  Pull-based processors
	// will call Collect() when a pull request arrives.
	Accumulator struct {
		// current maps `mapkey` to *record.
		current sync.Map

		// asyncInstruments is a set of
		// `*asyncInstrument` instances
		asyncLock        sync.Mutex
		asyncInstruments *internal.AsyncInstrumentState

		// currentEpoch is the current epoch number. It is
		// incremented in `Collect()`.
		currentEpoch int64

		// processor is the configured processor+configuration.
		processor export.Processor

		// collectLock prevents simultaneous calls to Collect().
		collectLock sync.Mutex

		// asyncSortSlice has a single purpose - as a temporary
		// place for sorting during labels creation to avoid
		// allocation.  It is cleared after use.
		asyncSortSlice label.Sortable

		// resource is applied to all records in this Accumulator.
		resource *resource.Resource

		// metricsProcessors are applied to all records in this Accumulator
		metricsProcessors []MetricsProcessor
	}

	syncInstrument struct {
		instrument
	}

	// mapkey uniquely describes a metric instrument in terms of
	// its InstrumentID and the encoded form of its labels.
	mapkey struct {
		descriptor *otel.Descriptor
		ordered    label.Distinct
	}

	// record maintains the state of one metric instrument.  Due
	// the use of lock-free algorithms, there may be more than one
	// `record` in existence at a time, although at most one can
	// be referenced from the `Accumulator.current` map.
	record struct {
		// refMapped keeps track of refcounts and the mapping state to the
		// Accumulator.current map.
		refMapped refcountMapped

		// updateCount is incremented on every Update.
		updateCount int64

		// collectedCount is set to updateCount on collection,
		// supports checking for no updates during a round.
		collectedCount int64

		// storage is the stored label set for this record,
		// except in cases where a label set is shared due to
		// batch recording.
		storage label.Set

		// labels is the processed label set for this record.
		// this may refer to the `storage` field in another
		// record if this label set is shared resulting from
		// `RecordBatch`.
		labels *label.Set

		// sortSlice has a single purpose - as a temporary
		// place for sorting during labels creation to avoid
		// allocation.
		sortSlice label.Sortable

		// inst is a pointer to the corresponding instrument.
		inst *syncInstrument

		// current implements the actual RecordOne() API,
		// depending on the type of aggregation.  If nil, the
		// metric was disabled by the exporter.
		current    export.Aggregator
		checkpoint export.Aggregator
	}

	instrument struct {
		meter             *Accumulator
		descriptor        otel.Descriptor
		metricsProcessors []MetricsProcessor
	}

	asyncInstrument struct {
		instrument
		// recorders maps ordered labels to the pair of
		// labelset and recorder
		recorders map[label.Distinct]*labeledRecorder
	}

	labeledRecorder struct {
		observedEpoch int64
		labels        *label.Set
		observed      export.Aggregator
	}

	// MetricsProcessor can be provided as an config option when creating
	// a new push controller
	MetricsProcessor interface {
		// OnMetricRecorded is execute everytime a metric is recorded by
		// the sync instrument implementation of an Accumulator, it generally
		// provides ability to correlate the context with the metrics
		OnMetricRecorded(context.Context, *[]label.KeyValue)
	}
)

var (
	_ api.MeterImpl     = &Accumulator{}
	_ api.AsyncImpl     = &asyncInstrument{}
	_ api.SyncImpl      = &syncInstrument{}
	_ api.BoundSyncImpl = &record{}

	ErrUninitializedInstrument = fmt.Errorf("use of an uninitialized instrument")
)

func (inst *instrument) Descriptor() api.Descriptor {
	return inst.descriptor
}

func (inst *instrument) getMetricsProcessors() []MetricsProcessor {
	return inst.metricsProcessors
}

func (a *asyncInstrument) Implementation() interface{} {
	return a
}

func (s *syncInstrument) Implementation() interface{} {
	return s
}

func (a *asyncInstrument) observe(number api.Number, labels *label.Set) {
	if err := aggregator.RangeTest(number, &a.descriptor); err != nil {
		global.Handle(err)
		return
	}
	recorder := a.getRecorder(labels)
	if recorder == nil {
		// The instrument is disabled according to the
		// AggregatorSelector.
		return
	}
	if err := recorder.Update(context.Background(), number, &a.descriptor); err != nil {
		global.Handle(err)
		return
	}
}

func (a *asyncInstrument) getRecorder(labels *label.Set) export.Aggregator {
	lrec, ok := a.recorders[labels.Equivalent()]
	if ok {
		if lrec.observedEpoch == a.meter.currentEpoch {
			// last value wins for Observers, so if we see the same labels
			// in the current epoch, we replace the old recorder
			a.meter.processor.AggregatorFor(&a.descriptor, &lrec.observed)
		} else {
			lrec.observedEpoch = a.meter.currentEpoch
		}
		a.recorders[labels.Equivalent()] = lrec
		return lrec.observed
	}
	var rec export.Aggregator
	a.meter.processor.AggregatorFor(&a.descriptor, &rec)
	if a.recorders == nil {
		a.recorders = make(map[label.Distinct]*labeledRecorder)
	}
	// This may store nil recorder in the map, thus disabling the
	// asyncInstrument for the labelset for good. This is intentional,
	// but will be revisited later.
	a.recorders[labels.Equivalent()] = &labeledRecorder{
		observed:      rec,
		labels:        labels,
		observedEpoch: a.meter.currentEpoch,
	}
	return rec
}

// acquireHandle gets or creates a `*record` corresponding to `kvs`,
// the input labels.  The second argument `labels` is passed in to
// support re-use of the orderedLabels computed by a previous
// measurement in the same batch.   This performs two allocations
// in the common case.
func (s *syncInstrument) acquireHandle(kvs []label.KeyValue, labelPtr *label.Set) *record {
	var rec *record
	var equiv label.Distinct

	if labelPtr == nil {
		// This memory allocation may not be used, but it's
		// needed for the `sortSlice` field, to avoid an
		// allocation while sorting.
		rec = &record{}
		rec.storage = label.NewSetWithSortable(kvs, &rec.sortSlice)
		rec.labels = &rec.storage
		equiv = rec.storage.Equivalent()
	} else {
		equiv = labelPtr.Equivalent()
	}

	// Create lookup key for sync.Map (one allocation, as this
	// passes through an interface{})
	mk := mapkey{
		descriptor: &s.descriptor,
		ordered:    equiv,
	}

	if actual, ok := s.meter.current.Load(mk); ok {
		// Existing record case.
		existingRec := actual.(*record)
		if existingRec.refMapped.ref() {
			// At this moment it is guaranteed that the entry is in
			// the map and will not be removed.
			return existingRec
		}
		// This entry is no longer mapped, try to add a new entry.
	}

	if rec == nil {
		rec = &record{}
		rec.labels = labelPtr
	}
	rec.refMapped = refcountMapped{value: 2}
	rec.inst = s

	s.meter.processor.AggregatorFor(&s.descriptor, &rec.current, &rec.checkpoint)

	for {
		// Load/Store: there's a memory allocation to place `mk` into
		// an interface here.
		if actual, loaded := s.meter.current.LoadOrStore(mk, rec); loaded {
			// Existing record case. Cannot change rec here because if fail
			// will try to add rec again to avoid new allocations.
			oldRec := actual.(*record)
			if oldRec.refMapped.ref() {
				// At this moment it is guaranteed that the entry is in
				// the map and will not be removed.
				return oldRec
			}
			// This loaded entry is marked as unmapped (so Collect will remove
			// it from the map immediately), try again - this is a busy waiting
			// strategy to wait until Collect() removes this entry from the map.
			//
			// This can be improved by having a list of "Unmapped" entries for
			// one time only usages, OR we can make this a blocking path and use
			// a Mutex that protects the delete operation (delete only if the old
			// record is associated with the key).

			// Let collector get work done to remove the entry from the map.
			runtime.Gosched()
			continue
		}
		// The new entry was added to the map, good to go.
		return rec
	}
}

func (s *syncInstrument) Bind(kvs []label.KeyValue) api.BoundSyncImpl {
	return s.acquireHandle(kvs, nil)
}

func (s *syncInstrument) RecordOne(ctx context.Context, number api.Number, kvs []label.KeyValue) {
		processor.OnMetricRecorded(ctx, &kvs)
	}
	h := s.acquireHandle(kvs, nil)
	defer h.Unbind()
	h.RecordOne(ctx, number)
}

// NewAccumulator constructs a new Accumulator for the given
// processor.  This Accumulator supports only a single processor.
//
// The Accumulator does not start any background process to collect itself
// periodically, this responsbility lies with the processor, typically,
// depending on the type of export.  For example, a pull-based
// processor will call Collect() when it receives a request to scrape
// current metric values.  A push-based processor should configure its
// own periodic collection.
func NewAccumulator(processor export.Processor, opts ...Option) *Accumulator {
	c := &Config{}
	for _, opt := range opts {
		opt.Apply(c)
	}

	return &Accumulator{
		asyncInstruments:  internal.NewAsyncInstrumentState(),
		processor:         processor,
		resource:          c.Resource,
		metricsProcessors: c.MetricsProcessors,
	}
}

// NewSyncInstrument implements api.MetricImpl.
func (m *Accumulator) NewSyncInstrument(descriptor api.Descriptor) (api.SyncImpl, error) {
	return &syncInstrument{
		instrument: instrument{
			descriptor:        descriptor,
			meter:             m,
			metricsProcessors: m.metricsProcessors,
		},
	}, nil
}

// NewAsyncInstrument implements api.MetricImpl.
func (m *Accumulator) NewAsyncInstrument(descriptor api.Descriptor, runner otel.AsyncRunner) (api.AsyncImpl, error) {
	a := &asyncInstrument{
		instrument: instrument{
			descriptor: descriptor,
			meter:      m,
		},
	}
	m.asyncLock.Lock()
	defer m.asyncLock.Unlock()
	m.asyncInstruments.Register(a, runner)
	return a, nil
}

// Collect traverses the list of active records and observers and
// exports data for each active instrument.  Collect() may not be
// called concurrently.
//
// During the collection pass, the export.Processor will receive
// one Export() call per current aggregation.
//
// Returns the number of records that were checkpointed.
func (m *Accumulator) Collect(ctx context.Context) int {
	m.collectLock.Lock()
	defer m.collectLock.Unlock()

	checkpointed := m.observeAsyncInstruments(ctx)
	checkpointed += m.collectSyncInstruments()
	m.currentEpoch++

	return checkpointed
}

func (m *Accumulator) collectSyncInstruments() int {
	checkpointed := 0

	m.current.Range(func(key interface{}, value interface{}) bool {
		// Note: always continue to iterate over the entire
		// map by returning `true` in this function.
		inuse := value.(*record)

		mods := atomic.LoadInt64(&inuse.updateCount)
		coll := inuse.collectedCount

		if mods != coll {
			// Updates happened in this interval,
			// checkpoint and continue.
			checkpointed += m.checkpointRecord(inuse)
			inuse.collectedCount = mods
			return true
		}

		// Having no updates since last collection, try to unmap:
		if unmapped := inuse.refMapped.tryUnmap(); !unmapped {
			// The record is referenced by a binding, continue.
			return true
		}

		// If any other goroutines are now trying to re-insert this
		// entry in the map, they are busy calling Gosched() awaiting
		// this deletion:
		m.current.Delete(inuse.mapkey())

		// There's a potential race between `LoadInt64` and
		// `tryUnmap` in this function.  Since this is the
		// last we'll see of this record, checkpoint
		mods = atomic.LoadInt64(&inuse.updateCount)
		if mods != coll {
			checkpointed += m.checkpointRecord(inuse)
		}
		return true
	})

	return checkpointed
}

// CollectAsync implements internal.AsyncCollector.
func (m *Accumulator) CollectAsync(kv []label.KeyValue, obs ...otel.Observation) {
	labels := label.NewSetWithSortable(kv, &m.asyncSortSlice)

	for _, ob := range obs {
		if a := m.fromAsync(ob.AsyncImpl()); a != nil {
			a.observe(ob.Number(), &labels)
		}
	}
}

func (m *Accumulator) observeAsyncInstruments(ctx context.Context) int {
	m.asyncLock.Lock()
	defer m.asyncLock.Unlock()

	asyncCollected := 0

	// TODO: change this to `ctx` (in a separate PR, with tests)
	m.asyncInstruments.Run(context.Background(), m)

	for _, inst := range m.asyncInstruments.Instruments() {
		if a := m.fromAsync(inst); a != nil {
			asyncCollected += m.checkpointAsync(a)
		}
	}

	return asyncCollected
}

func (m *Accumulator) checkpointRecord(r *record) int {
	if r.current == nil {
		return 0
	}
	err := r.current.SynchronizedMove(r.checkpoint, &r.inst.descriptor)
	if err != nil {
		global.Handle(err)
		return 0
	}

	a := export.NewAccumulation(&r.inst.descriptor, r.labels, m.resource, r.checkpoint)
	err = m.processor.Process(a)
	if err != nil {
		global.Handle(err)
	}
	return 1
}

func (m *Accumulator) checkpointAsync(a *asyncInstrument) int {
	if len(a.recorders) == 0 {
		return 0
	}
	checkpointed := 0
	for encodedLabels, lrec := range a.recorders {
		lrec := lrec
		epochDiff := m.currentEpoch - lrec.observedEpoch
		if epochDiff == 0 {
			if lrec.observed != nil {
				a := export.NewAccumulation(&a.descriptor, lrec.labels, m.resource, lrec.observed)
				err := m.processor.Process(a)
				if err != nil {
					global.Handle(err)
				}
				checkpointed++
			}
		} else if epochDiff > 1 {
			// This is second collection cycle with no
			// observations for this labelset. Remove the
			// recorder.
			delete(a.recorders, encodedLabels)
		}
	}
	if len(a.recorders) == 0 {
		a.recorders = nil
	}
	return checkpointed
}

// RecordBatch enters a batch of metric events.
func (m *Accumulator) RecordBatch(ctx context.Context, kvs []label.KeyValue, measurements ...api.Measurement) {
	// Labels will be computed the first time acquireHandle is
	// called.  Subsequent calls to acquireHandle will re-use the
	// previously computed value instead of recomputing the
	// ordered labels.
	var labelsPtr *label.Set
	for i, meas := range measurements {
		s := m.fromSync(meas.SyncImpl())
		if s == nil {
			continue
		}
		h := s.acquireHandle(kvs, labelsPtr)

		// Re-use labels for the next measurement.
		if i == 0 {
			labelsPtr = h.labels
		}

		defer h.Unbind()
		h.RecordOne(ctx, meas.Number())
	}
}

// RecordOne implements api.SyncImpl.
func (r *record) RecordOne(ctx context.Context, number api.Number) {
	if r.current == nil {
		// The instrument is disabled according to the AggregatorSelector.
		return
	}
	if err := aggregator.RangeTest(number, &r.inst.descriptor); err != nil {
		global.Handle(err)
		return
	}
	if err := r.current.Update(ctx, number, &r.inst.descriptor); err != nil {
		global.Handle(err)
		return
	}
	// Record was modified, inform the Collect() that things need
	// to be collected while the record is still mapped.
	atomic.AddInt64(&r.updateCount, 1)
}

// Unbind implements api.SyncImpl.
func (r *record) Unbind() {
	r.refMapped.unref()
}

func (r *record) mapkey() mapkey {
	return mapkey{
		descriptor: &r.inst.descriptor,
		ordered:    r.labels.Equivalent(),
	}
}

// fromSync gets a sync implementation object, checking for
// uninitialized instruments and instruments created by another SDK.
func (m *Accumulator) fromSync(sync otel.SyncImpl) *syncInstrument {
	if sync != nil {
		if inst, ok := sync.Implementation().(*syncInstrument); ok {
			return inst
		}
	}
	global.Handle(ErrUninitializedInstrument)
	return nil
}

// fromSync gets an async implementation object, checking for
// uninitialized instruments and instruments created by another SDK.
func (m *Accumulator) fromAsync(async otel.AsyncImpl) *asyncInstrument {
	if async != nil {
		if inst, ok := async.Implementation().(*asyncInstrument); ok {
			return inst
		}
	}
	global.Handle(ErrUninitializedInstrument)
	return nil
}
