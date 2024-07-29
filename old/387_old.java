        return record;
    }

    public Schema getSchema(String topic) {
        Schema schema = schemas.get(topic);
        if (schema == null) {
            throw new IllegalStateException("Avro schema not found for topic " + topic);
        }
        return schema;
    }
}
