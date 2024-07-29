				fmt.Printf("%s %s (%s)\r\n", name, appVersion, util.GetPlatformName())
				return
			}

			run()
		},
	}

	c.SetOutput(os.Stdout)
	c.PersistentFlags().StringVar(conf, "config", "node.yaml", "Path to the Node config file")
	c.PersistentFlags().BoolVar(version, "version", false, "Show Node version and exit")

	return c
}
