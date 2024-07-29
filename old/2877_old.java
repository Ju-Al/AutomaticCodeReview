
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
