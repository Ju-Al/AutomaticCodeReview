	return c.initialReq, true
}

func (c *committer) validateInitialResponse(response interface{}) error {
	commitResponse, _ := response.(*pb.StreamingCommitCursorResponse)
	if commitResponse.GetInitial() == nil {
		return errInvalidInitialCommitResponse
	}
	return nil
}

func (c *committer) onStreamStatusChange(status streamStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch status {
	case streamConnected:
		c.unsafeUpdateStatus(serviceActive, nil)
		// Once the stream connects, clear unconfirmed commits and immediately send
		// the latest desired commit offset.
		c.cursorTracker.ClearPending()
		c.unsafeCommitOffsetToStream()
		c.pollCommits.Start()

	case streamReconnecting:
		c.pollCommits.Stop()

	case streamTerminated:
		c.unsafeInitiateShutdown(serviceTerminated, c.stream.Error())
	}
}

func (c *committer) onResponse(response interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If an inconsistency is detected in the server's responses, immediately
	// terminate the committer, as correct processing of commits cannot be
	// guaranteed.
	processResponse := func() error {
		commitResponse, _ := response.(*pb.StreamingCommitCursorResponse)
		if commitResponse.GetCommit() == nil {
			return errInvalidCommitResponse
		}
		numAcked := commitResponse.GetCommit().GetAcknowledgedCommits()
		if numAcked <= 0 {
			return fmt.Errorf("pubsublite: server acknowledged an invalid commit count: %d", numAcked)
		}
		if err := c.cursorTracker.ConfirmOffsets(numAcked); err != nil {
			return err
		}
		c.unsafeCheckDone()
		return nil
	}
	if err := processResponse(); err != nil {
		c.unsafeInitiateShutdown(serviceTerminated, err)
	}
}

// commitOffsetToStream is called by the periodic background task.
func (c *committer) commitOffsetToStream() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.unsafeCommitOffsetToStream()
}

func (c *committer) unsafeCommitOffsetToStream() {
	nextOffset := c.cursorTracker.NextOffset()
	if nextOffset == nilCursorOffset {
		return
	}

	req := &pb.StreamingCommitCursorRequest{
		Request: &pb.StreamingCommitCursorRequest_Commit{
			Commit: &pb.SequencedCommitCursorRequest{
				Cursor: &pb.Cursor{Offset: nextOffset},
			},
		},
	}
	if c.stream.Send(req) {
		c.cursorTracker.AddPending(nextOffset)
	}
}

func (c *committer) unsafeInitiateShutdown(targetStatus serviceStatus, err error) {
	if !c.unsafeUpdateStatus(targetStatus, err) {
		return
	}

	// If it's a graceful shutdown, expedite sending final commits to the stream.
	if targetStatus == serviceTerminating {
		c.unsafeCommitOffsetToStream()
		c.unsafeCheckDone()
		return
	}
	// Otherwise immediately terminate the stream.
	c.unsafeTerminate()
}

func (c *committer) unsafeCheckDone() {
	// If the user stops the subscriber, they will no longer receive messages, but
	// the commit stream remains open to process acks for outstanding messages.
	if c.status == serviceTerminating && c.cursorTracker.Done() {
		c.unsafeTerminate()
	}
}

func (c *committer) unsafeTerminate() {
	c.acks.Release()
	c.pollCommits.Stop()
	c.stream.Stop()
}
