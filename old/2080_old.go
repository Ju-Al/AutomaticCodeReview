	return 0
}

func (c *OnDemandRunnerInspectCommand) Flags() *flag.Sets {
	return c.flagSet(0, nil)
}

func (c *OnDemandRunnerInspectCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *OnDemandRunnerInspectCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *OnDemandRunnerInspectCommand) Synopsis() string {
	return "Show detailed information about a configured auth method"
}

func (c *OnDemandRunnerInspectCommand) Help() string {
	return formatHelp(`
Usage: waypoint runner on-demand inspect NAME

  Show detailed information about an on-demand runner configuration.

`)
}
