    >>> for subgraph_batch in dataloader:
    ...     train_on(subgraph_batch)
    """
    def __init__(self, g, num_partitions, cache_directory, refresh=False):
        super().__init__(g)

        # First see if the cache is already there.  If so, directly read from cache.
        if not refresh and self._load_parts(cache_directory):
            return

        # Otherwise, build the cache.
        assignment = F.asnumpy(metis_partition_assignment(g, num_partitions))
        self._save_parts(assignment, cache_directory)

    def _cache_file_path(self, cache_directory):
        return os.path.join(cache_directory, 'cluster_gcn_cache')

    def _load_parts(self, cache_directory):
        path = self._cache_file_path(cache_directory)
        if not os.path.exists(path):
            return False

        with open(path, 'rb') as file_:
            self.part_indptr, self.part_indices = pickle.load(file_)
        return True

    def _save_parts(self, assignment, cache_directory):
        os.makedirs(cache_directory, exist_ok=True)

        self.part_indices = np.argsort(assignment)
        num_nodes_per_part = np.bincount(assignment)
        self.part_indptr = np.insert(np.cumsum(num_nodes_per_part), 0, 0)

        with open(self._cache_file_path(cache_directory), 'wb') as f:
            pickle.dump((self.part_indptr, self.part_indices), f)

    def __len__(self):
        return self.part_indptr.shape[0] - 1

    def __getitem__(self, i):
        nodes = self.part_indices[self.part_indptr[i]:self.part_indptr[i+1]]
        return self.g.subgraph(nodes)
