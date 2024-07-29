
    protected final Properties properties = new Properties();

    /**
     * @param name           name of the source, needs to be unique,
     *                       will be passed to the underlying Kafka
     *                       Connect source
     * @param connectorClass name of the Java class for the connector,
     *                       hardcoded for each type of DB
     */
    protected AbstractSourceBuilder(String name, String connectorClass) {
        properties.put("name", name);
        properties.put("connector.class", connectorClass);
        properties.put("database.history", HazelcastListDatabaseHistory.class.getName());
        properties.put("database.history.hazelcast.list.name", name);
        properties.put("tombstones.on.delete", "false");
    }

    /**
     * Can be used to set any property not covered by our builders,
     * or to override properties we have hidden.
     *
     * @param key   the name of the property to set
     * @param value the value of the property to set
     * @return the builder itself
     */
    public SELF setCustomProperty(String key, String value) {
        properties.put(key, value);
        return (SELF) this;
    }

    protected SELF setProperty(String key, String value) {
        properties.put(key, value);
        return (SELF) this;
    }

    protected SELF setProperty(String key, int value) {
        return setProperty(key, Integer.toString(value));
    }

    protected SELF setProperty(String key, boolean value) {
        return setProperty(key, Boolean.toString(value));
    }

    protected SELF setProperty(String key, String... values) {
        return setProperty(key, String.join(",", values));
    }

    protected static StreamSource<ChangeEvent> connect(@Nonnull Properties properties) {
        String name = properties.getProperty("name");
        FunctionEx<Processor.Context, CdcSource> createFn = ctx -> new CdcSource(ctx, properties);
        return SourceBuilder.timestampedStream(name, createFn)
                .fillBufferFn(CdcSource::fillBuffer)
                .createSnapshotFn(CdcSource::createSnapshot)
                .restoreSnapshotFn(CdcSource::restoreSnapshot)
                .destroyFn(CdcSource::destroy)
                .build();
    }

}
