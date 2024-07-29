
  private IcebergReaderFactory() {
  }

  public static CloseableIterable<Record> createReader(DataFile file, FileScanTask currentTask, InputFile inputFile,
      Schema tableSchema, boolean reuseContainers) {
    switch (file.format()) {
      case AVRO:
        return buildAvroReader(currentTask, inputFile, tableSchema, reuseContainers);
      case ORC:
        return buildOrcReader(currentTask, inputFile, tableSchema, reuseContainers);
      case PARQUET:
        return buildParquetReader(currentTask, inputFile, tableSchema, reuseContainers);

      default:
        throw new UnsupportedOperationException(String.format("Cannot read %s file: %s", file.format().name(),
            file.path()));
    }
  }

  private static CloseableIterable buildAvroReader(FileScanTask task, InputFile inputFile, Schema schema,
      boolean reuseContainers) {
    Avro.ReadBuilder builder = Avro.read(inputFile)
        .createReaderFunc(DataReader::create)
        .project(schema)
        .split(task.start(), task.length());

    if (reuseContainers) {
      builder.reuseContainers();
    }

    return builder.build();
  }

  private static CloseableIterable buildOrcReader(FileScanTask task, InputFile inputFile, Schema schema,
      boolean reuseContainers) {
    ORC.ReadBuilder builder = ORC.read(inputFile)
        .createReaderFunc(fileSchema -> GenericOrcReader.buildReader(schema, fileSchema))
        .project(schema)
        .split(task.start(), task.length());

    return builder.build();
  }

  private static CloseableIterable buildParquetReader(FileScanTask task, InputFile inputFile, Schema schema,
      boolean reuseContainers) {
    Parquet.ReadBuilder builder = Parquet.read(inputFile)
        .createReaderFunc(fileSchema  -> GenericParquetReaders.buildReader(schema, fileSchema))
        .project(schema)
        .split(task.start(), task.length());

    if (reuseContainers) {
      builder.reuseContainers();
    }

    return builder.build();
  }
}
