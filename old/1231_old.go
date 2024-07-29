}

var databaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Dump the given ledger tracker database",
	Long:  "Dump the given ledger tracker database",
	Args:  validateNoPosArgsFn,
	Run: func(cmd *cobra.Command, args []string) {
		if ledgerTrackerFilename == "" {
			cmd.HelpFunc()(cmd, args)
			return
		}
		outFile := os.Stdout
		var err error
		if outFileName != "" {
			outFile, err = os.OpenFile(outFileName, os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				reportErrorf("Unable to create file '%s' : %v", outFileName, err)
			}
		}
		err = printAccountsDatabase(ledgerTrackerFilename, ledger.CatchpointFileHeader{}, outFile)
		if err != nil {
			reportErrorf("Unable to print account database : %v", err)
		}
	},
}
