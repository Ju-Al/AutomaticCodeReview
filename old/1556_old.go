
	// invoke f with our initial client index
	f(attempt, client)

	return retry.BackoffFunc(func() (time.Duration, bool) {
		l.Lock()
		defer l.Unlock()
		attempt++

		// if we have reached maxConsecutiveFailures failures update client index
		if attempt%maxConsecutiveFailures == 0 {
			// check if we have reached our last client index
			if client == numOfClients {
				client = 0
			} else {
				client++
			}

			//invoke f with new client index
			f(attempt, client)
		}

		val, stop := next.Next()
		if stop {
			return 0, true
		}

		return val, false
	})
}
