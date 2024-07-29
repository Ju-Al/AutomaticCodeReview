		h.delegate.Handle(err)
		return
	}
	h.l.Print(err)
}

// Handler returns the global Handler instance. If no Handler instance has
// be explicitly set yet, a default Handler is returned that logs to STDERR
// until an Handler is set (all functionality is delegated to the set
// Handler once it is set).
func Handler() oterror.Handler {
	return globalHandler
}

// SetHandler sets the global Handler to be h.
func SetHandler(h oterror.Handler) {
	defaultHandler.setDelegate(h)
}

// Handle is a convience function for Handler().Handle(err)
func Handle(err error) {
	globalHandler.Handle(err)
}
