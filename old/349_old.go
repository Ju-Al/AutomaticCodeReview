
var useSynopsis = []string{"<public_key_as_appears_in_ls | public_key_id_as_appears_in_ls"}

type UseCmd struct{}

// Name is returns the name of the Dolt cli command. This is what is used on the command line to invoke the command
func (cmd UseCmd) Name() string {
	return "use"
}

// Description returns a description of the command
func (cmd UseCmd) Description() string {
	return useShortDesc
}

// CreateMarkdown creates a markdown file containing the helptext for the command at the given path
func (cmd UseCmd) CreateMarkdown(fs filesys.Filesys, path, commandStr string) error {
	ap := cmd.createArgParser()
	return cli.CreateMarkdown(fs, path, commandStr, useShortDesc, useLongDesc, useSynopsis, ap)
}

// RequiresRepo should return false if this interface is implemented, and the command does not have the requirement
// that it be run from within a data repository directory
func (cmd UseCmd) RequiresRepo() bool {
	return false
}

// EventType returns the type of the event to log
func (cmd UseCmd) EventType() eventsapi.ClientEventType {
	return eventsapi.ClientEventType_CREDS_USE
}

func (cmd UseCmd) createArgParser() *argparser.ArgParser {
	ap := argparser.NewArgParser()
	return ap
}

// Exec executes the command
func (cmd UseCmd) Exec(ctx context.Context, commandStr string, args []string, dEnv *env.DoltEnv) int {
	ap := cmd.createArgParser()
	help, usage := cli.HelpAndUsagePrinters(commandStr, useShortDesc, useLongDesc, useSynopsis, ap)
	apr := cli.ParseArgs(ap, args, help)
	args = apr.Args()
	if len(args) != 1 {
		return commands.HandleVErrAndExitCode(errhand.BuildDError("error: expected exactly one credential public key or key id as argument").Build(), usage)
	}

	credsDir, verr := actions.EnsureCredsDir(dEnv)

	if verr == nil {
		jwkFilePath, err := dEnv.FindCreds(credsDir, args[0])
		if err == nil {
			cred, err := creds.JWKCredsReadFromFile(dEnv.FS, jwkFilePath)
			if err != nil {
				verr = errhand.BuildDError("error: failed to read credential %s", args[0]).AddCause(err).Build()
			} else {
				gcfg, hasGCfg := dEnv.Config.GetConfig(env.GlobalConfig)
				if !hasGCfg {
					panic("global config not found.  Should create it here if this is a thing.")
				}
				err := gcfg.SetStrings(map[string]string{env.UserCreds: cred.KeyIDBase32Str()})
				if err != nil {
					verr = errhand.BuildDError("error: updating user credentials in config").AddCause(err).Build()
				}
			}
		} else {
			verr = errhand.BuildDError("error: failed to find credential %s", args[0]).AddCause(err).Build()
		}
	}

	return commands.HandleVErrAndExitCode(verr, usage)
}
