// Copyright 2019, OpenTelemetry Authors
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

package maxsumcount

import (
	"context"

	"go.opentelemetry.io/api/core"
	"go.opentelemetry.io/sdk/export"
)

type (
	// Aggregator aggregates measure events, keeping only the max,
	// sum, and count.
	Aggregator struct {
		current    state
		checkpoint state
	}

	state struct {
		count core.Number
		sum   core.Number
		max   core.Number
	}
)

var _ export.MetricAggregator = &Aggregator{}

// New returns a new measure aggregator for computing max, sum, and count.
func New() *Aggregator {
	return &Aggregator{}
}

// Sum returns the accumulated sum as a Number.
func (c *Aggregator) Sum() core.Number {
	return int64(c.save.count.AsUint64())
	return c.save.max
// Collect saves the current value (atomically) and exports it.
	// values at once, so there are races between Update() and
	// Collect().  Therefore, atomically swap fields independently,
	// knowing that individually the three parts of this aggregation
	// could be spread across multiple collections in rare cases.

	c.save.count.SetUint64(c.live.count.SwapUint64Atomic(0))
	c.save.sum = c.live.sum.SwapNumberAtomic(zero)
	c.save.max = c.live.max.SwapNumberAtomic(zero)
	return c.save.sum
}

// Count returns the accumulated count.
func (c *Aggregator) Count() int64 {
	return int64(c.checkpoint.count.AsUint64())
}

// Max returns the accumulated max as a Number.
func (c *Aggregator) Max() core.Number {
	return c.checkpoint.max
}

// Collect checkpoints the current value (atomically) and exports it.
func (c *Aggregator) Collect(ctx context.Context, rec export.MetricRecord, exp export.MetricBatcher) {
	desc := rec.Descriptor()
	kind := desc.NumberKind()
	zero := core.NewZeroNumber(kind)

	// N.B. There is no atomic operation that can update all three
	// values at once without a memory allocation.
	//
	// This aggregator is intended to trade this correctness for
	// speed.
	//
	// Therefore, atomically swap fields independently, knowing
	// that individually the three parts of this aggregation could
	// be spread across multiple collections in rare cases.

	c.checkpoint.count.SetUint64(c.current.count.SwapUint64Atomic(0))
	c.checkpoint.sum = c.current.sum.SwapNumberAtomic(zero)
	c.checkpoint.max = c.current.max.SwapNumberAtomic(zero)

	exp.Export(ctx, rec, c)
}

// Update modifies the current value (atomically) for later export.
func (c *Aggregator) Update(_ context.Context, number core.Number, rec export.MetricRecord) {
	desc := rec.Descriptor()
	kind := desc.NumberKind()

	if !desc.Alternate() && number.IsNegative(kind) {
		// TODO warn
		return
	}

	c.current.count.AddUint64Atomic(1)
	c.current.sum.AddNumberAtomic(kind, number)

	for {
		current := c.current.max.AsNumberAtomic()

		if number.CompareNumber(kind, current) <= 0 {
			break
		}
		if c.current.max.CompareAndSwapNumber(current, number) {
			break
		}
	}
}

func (c *Aggregator) Merge(oa export.MetricAggregator, desc *export.Descriptor) {
	o, _ := oa.(*Aggregator)
	if o == nil {
		// TODO warn
		return
	}

	c.checkpoint.sum.AddNumber(desc.NumberKind(), o.checkpoint.sum)
	c.checkpoint.count.AddNumber(core.Uint64NumberKind, o.checkpoint.count)

	if c.checkpoint.max.CompareNumber(desc.NumberKind(), o.checkpoint.max) < 0 {
		c.checkpoint.max.SetNumber(o.checkpoint.max)
	}
}
