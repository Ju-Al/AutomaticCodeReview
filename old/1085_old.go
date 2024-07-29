					c.UI.Error(fmt.Sprintf("Error saving auth token to %q keyring: %s", keyringType, err))
					gotErr = true
				}

			default:
				krConfig := nkeyring.Config{
					LibSecretCollectionName: "login",
					PassPrefix:              "HashiCorp_Boundary",
					AllowedBackends:         []nkeyring.BackendType{nkeyring.BackendType(keyringType)},
				}

				kr, err := nkeyring.Open(krConfig)
				if err != nil {
					c.UI.Error(fmt.Sprintf("Error opening %q keyring: %s", keyringType, err))
					gotErr = true
					break
				}

				if err := kr.Set(nkeyring.Item{
					Key:  tokenName,
					Data: []byte(base64.RawStdEncoding.EncodeToString(marshaled)),
				}); err != nil {
					c.UI.Error(fmt.Sprintf("Error storing token in %q keyring: %s", keyringType, err))
					gotErr = true
					break
				}
			}

			if !gotErr {
				c.UI.Output("\nThe token was successfully stored in the chosen keyring and is not displayed here.")
			}
		}
	}

	switch {
	case gotErr:
		c.UI.Warn(fmt.Sprintf("The token was not successfully saved to a system keyring. The token is:\n\n%s\n\nIt must be manually passed in via the BOUNDARY_TOKEN env var or -token flag. Storing the token can also be disabled via -keyring-type=none.", token.Token))
	case c.FlagKeyringType == "none":
		c.UI.Warn("\nStoring the token in a keyring was disabled. The token is:")
		c.UI.Output(token.Token)
		c.UI.Warn("Please be sure to store it safely!")
	}

	return base.CommandSuccess
}
