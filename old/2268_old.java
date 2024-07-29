  }

  @Override
  public IteratorUtil.IteratorScope getIteratorScope() {
    return IteratorUtil.IteratorScope.majc;
  }

  @Override
  public RateLimiter getReadLimiter() {
    return readLimiter;
  }

  @Override
  public RateLimiter getWriteLimiter() {
    return writeLimiter;
  }

  @Override
  public SystemIteratorEnvironment createIteratorEnv(ServerContext context,
      AccumuloConfiguration acuTableConf, TableId tableId) {
    return new TabletIteratorEnvironment(context, IteratorUtil.IteratorScope.majc,
        !propagateDeletes, acuTableConf, tableId, kind);
  }

  @Override
  public SortedKeyValueIterator<Key,Value> getMinCIterator() {
    throw new UnsupportedOperationException();
  }

  @Override
  public TCompactionReason getReason() {
    switch (kind) {
      case USER:
        return TCompactionReason.USER;
      case CHOP:
        return TCompactionReason.CHOP;
      case SELECTOR:
      case SYSTEM:
      default:
        return TCompactionReason.SYSTEM;
    }
  }
}
