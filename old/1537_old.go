		}
	}

	newCheck := check{
		name:       name,
		expression: expression,
		enforced:   enforce,
	}
	c.checks = append(c.checks, newCheck)

	return newCheck, nil
}

func (c *checkCollection) DropCheck(name string) error {
	for i, chk := range c.checks {
		if strings.ToLower(name) == strings.ToLower(chk.name) {
			c.checks = append(c.checks[:i], c.checks[i+1:]...)
			return nil
		}
	}
	return nil
}

func (c *checkCollection) AllChecks() []Check {
	checks := make([]Check, len(c.checks))
	for i, check := range c.checks {
		checks[i] = check
	}
	return checks
}

func (c *checkCollection) Count() int {
	return len(c.checks)
}

func NewCheckCollection() CheckCollection {
	return &checkCollection{
		checks: make([]check, 0),
	}
}
