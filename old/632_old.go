	}

	return nil
}

func (c *AssignmentCollector) putIncorporatedAtHeight(incorporatdAtHeight uint64, incorporatedBlockID flow.Identifier) {
	c.incorporatedAtHeightLock.Lock()
	defer c.incorporatedAtHeightLock.Unlock()
	ids, ok := c.incorporatedAtHeight[incorporatdAtHeight]
	if !ok {
		ids = make([]flow.Identifier, 0)
	}
	ids = append(ids, incorporatedBlockID)
	c.incorporatedAtHeight[incorporatdAtHeight] = ids
}

func (c *AssignmentCollector) ProcessIncorporatedResult(incorporatedResult *flow.IncorporatedResult) error {
	incorporatedBlockID := incorporatedResult.IncorporatedBlockID
	if collector := c.collectorByBlockID(incorporatedBlockID); collector != nil {
		return nil
	}

	// chunk assigment is based on the first block in the fork that incorporates the result
	assignment, err := c.assigner.Assign(incorporatedResult.Result, incorporatedBlockID)
	if err != nil {
		return engine.NewInvalidInputErrorf("could not determine chunk assignment: %w", err)
	}

	// pre-select all authorized verifiers at the block that is being sealed
	c.authorizedApprovers, err = c.authorizedVerifiersAtBlock(incorporatedResult.Result.BlockID)
	if err != nil {
		return engine.NewInvalidInputErrorf("could not determine authorized verifiers for sealing candidate: %w", err)
	}

	incorporatedAtHeight, err := c.getCachedBlockHeight(incorporatedBlockID)
	if err != nil {
		return fmt.Errorf("coulld not determine height of incorporated block %s: %w",
			incorporatedBlockID, err)
	}

	collector := NewApprovalCollector(incorporatedResult, assignment, c.seals, c.requiredApprovalsForSealConstruction)

	c.putCollector(incorporatedBlockID, collector)
	c.putIncorporatedAtHeight(incorporatedAtHeight, incorporatedBlockID)

	// process approvals that have passed needed checks and are ready to be processed
	for _, approvalID := range c.verifiedApprovalsCache.Ids() {
		if approval := c.verifiedApprovalsCache.Peek(approvalID); approval != nil {
			// those approvals are verified already and shouldn't yield any errors
			_ = collector.ProcessApproval(approval)
		}
	}

	return nil
}

// OnBlockFinalizedAtHeight is responsible to cleanup collectors that are incorporated in orphan blocks
// by passing blockID and height of last finalized block we are able to identify if collector is for valid chain or
// for orphan fork.
func (c *AssignmentCollector) OnBlockFinalizedAtHeight(blockID flow.Identifier, blockHeight uint64) {
	c.incorporatedAtHeightLock.Lock()
	ids, ok := c.incorporatedAtHeight[blockHeight]
	if ok {
		delete(c.incorporatedAtHeight, blockHeight)
	}
	c.incorporatedAtHeightLock.Unlock()

	orphanBlockIds := make([]flow.Identifier, 0)
	// collect orphan incorporated blocks
	for _, id := range ids {
		if id != blockID {
			orphanBlockIds = append(orphanBlockIds, id)
		}
	}

	c.eraseCollectors(orphanBlockIds)
}

func (c *AssignmentCollector) putCollector(incorporatedBlockID flow.Identifier, collector *ApprovalCollector) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.collectors[incorporatedBlockID] = collector
}

func (c *AssignmentCollector) eraseCollectors(incorporatedBlockIds []flow.Identifier) {
	if len(incorporatedBlockIds) == 0 {
		return
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	for _, incorporatedBlockID := range incorporatedBlockIds {
		delete(c.collectors, incorporatedBlockID)
	}
}

func (c *AssignmentCollector) allCollectors() []*ApprovalCollector {
	c.lock.RLock()
	defer c.lock.RUnlock()
	collectors := make([]*ApprovalCollector, 0, len(c.collectors))
	for _, collector := range c.collectors {
		collectors = append(collectors, collector)
	}
	return collectors
}

func (c *AssignmentCollector) verifySignature(approval *flow.ResultApproval, nodeIdentity *flow.Identity) error {
	id := approval.Body.ID()
	valid, err := c.verifier.Verify(id[:], approval.VerifierSignature, nodeIdentity.StakingPubKey)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}

	if !valid {
		return engine.NewInvalidInputErrorf("invalid signature for (%x)", nodeIdentity.NodeID)
	}

	return nil
}

// validateApproval performs result level checks of flow.ResultApproval
// checks:
// 	verification node identity
//  signature of verification node
// returns nil on successful check
func (c *AssignmentCollector) validateApproval(approval *flow.ResultApproval) error {
	identity, found := c.authorizedApprovers[approval.Body.ApproverID]
	if !found {
		return engine.NewInvalidInputErrorf("approval not from authorized verifier")
	}

	err := c.verifySignature(approval, identity)
	if err != nil {
		return fmt.Errorf("invalid approval signature: %w", err)
	}

	return nil
}

// validateAndCache performs validation of approval and saves it into cache
// expects that execution result was discovered before calling this function
func (c *AssignmentCollector) validateAndCache(approval *flow.ResultApproval) error {
	err := c.validateApproval(approval)
	if err != nil {
		return fmt.Errorf("could not validate approval: %w", err)
	}

	c.verifiedApprovalsCache.Put(approval)
	return nil
}

func (c *AssignmentCollector) ProcessAssignment(approval *flow.ResultApproval) error {
	err := c.validateAndCache(approval)
	if err != nil {
		return fmt.Errorf("could not validate and cache approval: %w", err)
	}

	for _, collector := range c.allCollectors() {
		err := collector.ProcessApproval(approval)
		if err != nil {
			return fmt.Errorf("could not process assignment for collector %v: %w", collector.incorporatedResult.IncorporatedBlockID, err)
		}
	}

	return nil
}

func (c *AssignmentCollector) RequestMissingApprovals(maxHeightForRequesting uint64) error {
	for _, collector := range c.allCollectors() {
		// not finding the block that the result was incorporated in is a fatal
		// error at this stage
		blockHeight, err := c.getCachedBlockHeight(collector.IncorporatedBlockID)
		if err != nil {
			return fmt.Errorf("could not retrieve block: %w", err)
		}

		if blockHeight > maxHeightForRequesting {
			continue
		}

		// If we got this far, height `block.Height` must be finalized, because
		// maxHeightForRequesting is lower than the finalized height.

		// Skip result if it is incorporated in a block that is _not_ part of
		// the finalized fork.
		//finalizedBlockAtHeight, err := c.state.AtHeight(block.Height).Head()
		//if err != nil {
		//	return fmt.Errorf("could not retrieve finalized block for finalized height %d: %w", block.Height, err)
		//}
		// TODO: replace this check with cleanup while processing finalized blocks in approval processing core
		//if finalizedBlockAtHeight.ID() != collector.IncorporatedBlockID {
		//	// block is in an orphaned fork
		//	continue
		//}

		for chunkIndex, verifiers := range collector.CollectMissingVerifiers() {
			// Retrieve information about requests made for this chunk. Skip
			// requesting if the blackout period hasn't expired. Otherwise,
			// update request count and reset blackout period.
			requestTrackerItem := c.requestTracker.Get(c.ResultID, collector.IncorporatedBlockID, chunkIndex)
			if requestTrackerItem.IsBlackout() {
				continue
			}
			requestTrackerItem.Update()

			// for monitoring/debugging purposes, log requests if we start
			// making more than 10
			if requestTrackerItem.Requests >= 10 {
				log.Debug().Msgf("requesting approvals for result %v, incorporatedBlockID %v chunk %d: %d requests",
					c.ResultID,
					collector.IncorporatedBlockID,
					chunkIndex,
					requestTrackerItem.Requests,
				)
			}

			// prepare the request
			req := &messages.ApprovalRequest{
				Nonce:      rand.Uint64(),
				ResultID:   c.ResultID,
				ChunkIndex: chunkIndex,
			}

			err = c.approvalConduit.Publish(req, verifiers...)
			if err != nil {
				log.Error().Err(err).
					Msgf("could not publish approval request for chunk %d", chunkIndex)
			}
		}
	}
	return nil
}
