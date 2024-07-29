					"phonetool").
					Return("test", nil)

				return &envUpgradeOpts{
					envUpgradeVars: envUpgradeVars{
						appName: "phonetool",
					},
					sel: m,
				}
			},
			wantedAppName: "phonetool",
			wantedEnvName: "test",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			opts := tc.given(ctrl)

			err := opts.Ask()

			if tc.wantedErr != nil {
				require.EqualError(t, err, tc.wantedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantedAppName, opts.appName)
				require.Equal(t, tc.wantedEnvName, opts.name)
			}
		})
	}
}
