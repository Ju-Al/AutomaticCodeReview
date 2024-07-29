    private Long snapshotId;
    private Expression rowFilter;
    private boolean ignoreResiduals;
    private boolean caseSensitive;
    private boolean colStats;
    private Collection<String> selectedColumns;
    private ImmutableMap<String, String> options;
    private Long fromSnapshotId;
    private Long toSnapshotId;

    private Builder() {
      this.snapshotId = null;
      this.rowFilter = Expressions.alwaysTrue();
      this.ignoreResiduals = false;
      this.caseSensitive = false;
      this.colStats = false;
      this.selectedColumns = null;
      this.options = ImmutableMap.of();
      this.fromSnapshotId = null;
      this.toSnapshotId = null;
    }

    private Builder(TableScanContext other) {
      this.snapshotId = other.snapshotId;
      this.rowFilter = other.rowFilter;
      this.ignoreResiduals = other.ignoreResiduals;
      this.caseSensitive = other.caseSensitive;
      this.colStats = other.colStats;
      this.selectedColumns = other.selectedColumns;
      this.options = ImmutableMap.copyOf(other.options);
      this.fromSnapshotId = other.fromSnapshotId;
      this.toSnapshotId = other.toSnapshotId;
    }

    Builder snapshotId(Long id) {
      this.snapshotId = id;
      return this;
    }

    Builder rowFilter(Expression filter) {
      this.rowFilter = filter;
      return this;
    }

    Builder ignoreResiduals(boolean shouldIgnoreResiduals) {
      this.ignoreResiduals = shouldIgnoreResiduals;
      return this;
    }

    Builder caseSensitive(boolean isCaseSensitive) {
      this.caseSensitive = isCaseSensitive;
      return this;
    }

    Builder colStats(boolean shouldUseColumnStats) {
      this.colStats = shouldUseColumnStats;
      return this;
    }

    Builder selectedColumns(Collection<String> columns) {
      this.selectedColumns = columns;
      return this;
    }

    Builder options(Map<String, String> extraOptions) {
      this.options = ImmutableMap.copyOf(extraOptions);
      return this;
    }

    Builder fromSnapshotId(Long id) {
      this.fromSnapshotId = id;
      return this;
    }

    Builder toSnapshotId(Long id) {
      this.toSnapshotId = id;
      return this;
    }

    TableScanContext build() {
      return new TableScanContext(snapshotId, rowFilter, ignoreResiduals, caseSensitive, colStats,
          selectedColumns, options, fromSnapshotId, toSnapshotId);
    }
  }
}
