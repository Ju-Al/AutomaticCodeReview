/*
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package com.hazelcast.jet.sql.impl.aggregate;

import com.hazelcast.nio.ObjectDataInput;
import com.hazelcast.nio.ObjectDataOutput;

import java.io.IOException;
import java.util.HashSet;
import java.util.Set;

class DistinctSqlAggregation implements SqlAggregation {

    // once aggregation is serialized, accumulate() is never used again
    private transient Set<Object> values;

    private SqlAggregation delegate;

    @SuppressWarnings("unused")
    private DistinctSqlAggregation() {
    }

    DistinctSqlAggregation(SqlAggregation delegate) {
        this.values = new HashSet<>();

        this.delegate = delegate;
    }

    @Override
    public void accumulate(Object value) {
        if (value != null && values != null && !values.add(value)) {
            return;
        }

        delegate.accumulate(value);
    }

    @Override
    public void combine(SqlAggregation other0) {
        DistinctSqlAggregation other = (DistinctSqlAggregation) other0;

        delegate.combine(other.delegate);
    }

    @Override
    public Object collect() {
        return delegate.collect();
    }

    @Override
    public void writeData(ObjectDataOutput out) throws IOException {
        out.writeObject(delegate);
    }

    @Override
    public void readData(ObjectDataInput in) throws IOException {
        delegate = in.readObject();
    }
}