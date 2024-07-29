
		var iter *ctc.CanonicalTransactionChainTransactionEnqueuedIterator
		iter, err = s.contract.FilterTransactionEnqueued(&bind.FilterOpts{
			Start:   start,
			End:     &end,
			Context: ctxt,
		}, nil, nil, nil)
		if err != nil {
			fmt.Printf("Unable to query events for block range start=%d, end=%d; error=%v\n",
				start, end, err)
			continue
		}
		return iter, nil
	}
	return nil, err
}

func (s *Service) Update(start uint64, newHeader *types.Header) error {
	var lowest = BlockLocator{
		Number: s.cfg.StartBlockNumber,
		Hash:   common.HexToHash(s.cfg.StartBlockHash),
	}
	highestConfirmed, err := s.cfg.DB.GetHighestBlock()
	if err != nil {
		return err
	}
	if highestConfirmed != nil {
		lowest = *highestConfirmed
	}

	headers := s.headerSelector.NewHead(s.ctx, lowest.Number, newHeader, s.backend)
	if len(headers) == 0 {
		return errNoNewBlocks
	}

	if lowest.Number+1 != headers[0].Number.Uint64() {
		fmt.Printf("Block number of block=%d hash=%s does not "+
			"immediately follow lowest block=%d hash=%s\n",
			headers[0].Number.Uint64(), headers[0].Hash(),
			lowest.Number, lowest.Hash)
		return nil
	}

	if lowest.Hash != headers[0].ParentHash {
		fmt.Printf("Parent hash of block=%d hash=%s does not "+
			"connect to lowest block=%d hash=%s\n", headers[0].Number.Uint64(),
			headers[0].Hash(), lowest.Number, lowest.Hash)
		return nil
	}

	startHeight := headers[0].Number.Uint64()
	endHeight := headers[len(headers)-1].Number.Uint64()

	iter, err := s.fetchBlockEventIterator(startHeight, endHeight)
	if err != nil {
		return err
	}

	depositsByBlockhash := make(map[common.Hash][]Deposit)
	for iter.Next() {
		depositsByBlockhash[iter.Event.Raw.BlockHash] = append(
			depositsByBlockhash[iter.Event.Raw.BlockHash], Deposit{
				QueueIndex: iter.Event.QueueIndex.Uint64(),
				TxHash:     iter.Event.Raw.TxHash,
				L1TxOrigin: iter.Event.L1TxOrigin,
				Target:     iter.Event.Target,
				GasLimit:   iter.Event.GasLimit,
				Data:       iter.Event.Data,
			})
	}
	if err := iter.Error(); err != nil {
		return err
	}

	for _, header := range headers {
		blockHash := header.Hash()
		number := header.Number.Uint64()
		deposits := depositsByBlockhash[blockHash]

		block := &IndexedBlock{
			Hash:       blockHash,
			ParentHash: header.ParentHash,
			Number:     number,
			Timestamp:  header.Time,
			Deposits:   deposits,
		}

		err := s.cfg.DB.AddIndexedBlock(block)
		if err != nil {
			fmt.Printf("Unable to import block=%d hash=%s err=%v "+
				"block: %v\n", number, blockHash, err, block)
			return err
		}

		fmt.Printf("Import block=%d hash=%s with %d deposits\n",
			number, blockHash, len(block.Deposits))
		for _, deposit := range block.Deposits {
			fmt.Printf("Deposit: l1_tx_origin=%s target=%s "+
				"gas_limit=%d queue_index=%d\n", deposit.L1TxOrigin,
				deposit.Target, deposit.GasLimit, deposit.QueueIndex)
		}
	}

	latestHeaderNumber := headers[len(headers)-1].Number.Uint64()
	newHeaderNumber := newHeader.Number.Uint64()
	if latestHeaderNumber+s.cfg.ConfDepth-1 == newHeaderNumber {
		return errNoNewBlocks
	}
	return nil
}

func (s *Service) Start() error {
	if s.cfg.ChainID == nil {
		return errNoChainID
	}
	s.wg.Add(1)
	go s.Loop(context.Background())
	return nil
}

func (s *Service) Stop() error {
	s.cancel()
	s.wg.Wait()
	if err := s.cfg.DB.Close(); err != nil {
		return err
	}
	return nil
}
