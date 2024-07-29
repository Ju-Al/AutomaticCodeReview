}

func (rc *rexCache) Retrieve(target *core.BuildTarget, key []byte) bool {
	if err := rc.client.Retrieve(target, key); err != nil {
		if remote.IsNotFound(err) {
			log.Debug("Artifacts for %s [key %s] don't exist in remote cache", target.Label, hex.EncodeToString(key))
		} else {
			log.Warning("Error retrieving artifacts for %s from remote cache: %s", target.Label, err)
		}
		return false
	}
	return true
}

func (rc *rexCache) RetrieveExtra(target *core.BuildTarget, key []byte, file string) bool {
	return false
}

func (rc *rexCache) Clean(target *core.BuildTarget) {
	// There is no API for this, so we just don't do it.
}

func (rc *rexCache) CleanAll() {
	// Similarly here.
}

func (rc *rexCache) Shutdown() {
}
