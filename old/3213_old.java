          .commit();
      LOG.info("Compaction transaction has been committed successfully, total delete files: {}, total add files: {}",
          currentDataFiles, dataFiles);

    } catch (Exception e) {
      Tasks.foreach(Iterables.transform(dataFiles, f -> f.path().toString()))
          .noRetry()
          .suppressFailureWhenFinished()
          .onFailure((location, exc) -> LOG.warn("Failed to delete: {}", location, exc))
          .run(table.io()::deleteFile);
      throw e;
    }
  }

  /**
   * It is notified that no more data will arrive on the input.
   */
  @Override
  public void endInput() throws Exception {
  }

  @Override
  public void dispose() throws Exception {
    if (tableLoader != null) {
      tableLoader.close();
    }
  }

  private void emit(EndCheckpoint result) {
    output.collect(new StreamRecord<>(result));
  }
}
