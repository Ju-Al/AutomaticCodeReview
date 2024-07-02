/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package org.apache.iceberg;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import org.apache.hadoop.conf.Configuration;
import org.apache.iceberg.hadoop.HadoopTables;
import org.apache.iceberg.types.Types;
import org.apache.spark.sql.Dataset;
import org.apache.spark.sql.Row;
import org.apache.spark.sql.SparkSession;
import org.junit.After;
import org.junit.Before;
import org.junit.Test;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import static org.apache.iceberg.types.Types.NestedField.optional;

/**
 * This class tests how Iceberg handles concurrency when reading and writing at the same time
 */
public class ConcurrencyTest {

  private static final Logger log = LoggerFactory.getLogger(ConcurrencyTest.class);

  private Configuration conf = new Configuration();
  private Schema schema = new Schema(
      optional(1, "key", Types.LongType.get()),
      optional(2, "value", Types.StringType.get())
  );
  private SparkSession sparkReader;
  private SparkSession sparkWriter;
  private File tableLocation;
  private Table table;

  private List<SimpleRecord> data = new ArrayList<>();

  @Before
  public void before() throws IOException {
    tableLocation = Files.createTempDirectory("temp").toFile();

    sparkReader = SparkSession.builder().master("local[2]").getOrCreate();
    sparkWriter = SparkSession.builder().master("local[2]").getOrCreate();

    sparkReader.sparkContext().setLogLevel("WARN");
    sparkWriter.sparkContext().setLogLevel("WARN");

    HadoopTables tables = new HadoopTables(conf);
    table = tables.create(schema, tableLocation.toString());

    for (int i = 0; i < 1000000; i++) {
      data.add(new SimpleRecord(1, "bdp"));
    }

    log.info("End of setup phase");
  }

  /**
   * The test creates 500 read tasks and one really long write (writing 1 mil rows)
   * and uses threading to call the tasks concurrently.
   */
  @Test
  public void writingAndReadingConcurrently() throws InterruptedException {
    ExecutorService threadPool = Executors.newFixedThreadPool(5);
    List<Callable<Void>> tasks = new ArrayList<>();

    Callable<Void> write = () -> writeToTable(data);
    tasks.add(write);

    for (int i = 0; i < 500; i++) {
      Callable<Void> getReads = () -> readTable();
      tasks.add(getReads);
    }

    threadPool.invokeAll(tasks);
    threadPool.shutdown();

    table.refresh();
    readTable();
  }

  private Void readTable() {
    Dataset<Row> results = sparkReader.read()
        .format("iceberg")
        .load(tableLocation.toString());

    log.info("" + results.count());
    return null;
  }
  private Void writeToTable(List<SimpleRecord> writeData) {
    log.info("WRITING!");
    Dataset<Row> df = sparkWriter.createDataFrame(writeData, SimpleRecord.class);
    df.select("key",  "value").write()
        .format("iceberg")
        .mode("append")
        .save(tableLocation.toString());
    return null;
  }

  @After
  public void after() {
    sparkReader.stop();
    sparkWriter.stop();
  }
}
