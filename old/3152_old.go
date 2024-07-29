}

// Resume polling. The task is executed after the polling period.
func (pt *periodicTask) Resume() {
	pt.ticker.Reset(pt.period)
}

// Pause temporarily suspends the polling.
func (pt *periodicTask) Pause() {
	pt.ticker.Stop()
}

// Stop permanently stops the periodic task.
func (pt *periodicTask) Stop() {
	// Prevent a panic if the stop channel has already been closed.
	if !pt.stopped {
		close(pt.stop)
		pt.stopped = true
	}
}

func (pt *periodicTask) poll() {
	for {
		select {
		case <-pt.stop:
			// Ends the goroutine.
			return
		case <-pt.ticker.C:
			pt.task()
		}
	}
}
