package cli

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/balaji01-4d/pgxcli/internal/app"
	"github.com/balaji01-4d/pgxcli/internal/config"
	"github.com/balaji01-4d/pgxcli/internal/database"
	"github.com/balaji01-4d/pgxcli/internal/logger"
	"github.com/balaji01-4d/pgxcli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
	osUser  = osUsername()
)

func NewRootCmd(ctx context.Context, cliCtx *CliContext) *cobra.Command {
	var (
		debugFlag           debugFlag
		hostFlag            hostFlag
		portFlag            portFlag
		dbNameFlag          dbNameFlag
		usernameFlag        usernameFlag
		neverPromptFlag     neverPromptFlag
		forcePromptFlag     forcePromptFlag
		interactiveConnFlag interactiveConnFlag
	)

	var (
		finalDatabase string
		finalUser     string
		finalHost     string
		finalPort     uint16
		finalPassword string
	)

	rootCmd := &cobra.Command{
		Use:     "pgxcli [DBNAME] [USERNAME]",
		Short:   "Interactive PostgreSQL command-line client for querying and managing databases.",
		Version: version,
		Args:    cobra.MaximumNArgs(2), // Database name and username are optional example: pgxcli mydb myuser

		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			cliCtx.config = cfg
			if err := cliCtx.Printer.SetPagerMode(cfg.Main.Pager); err != nil {
				return err
			}

			logger, err := logger.InitLogger(bool(debugFlag), cfg.Main.LogFile)
			if err != nil {
				return err
			}

			cliCtx.Logger = logger

			return nil
		},

		PreRunE: func(cmd *cobra.Command, args []string) error {
			argDB, argUser := parsePositionalDBAndUser(args)

			if bool(interactiveConnFlag) {
				// In interactive mode, flags / args are used as defaults
				// User can overrid them in the form
				// Priority is flag, arg, env, default
				var formUser string
				var formHost, formPort string
				formDB := firstNonEmpty(string(dbNameFlag), argDB)
				// TODO: implement getDefaultDB

				formUser = firstNonEmpty(string(usernameFlag), argUser, getDefaultUser())

				if cmd.Flags().Changed("host") {
					formHost = string(hostFlag)
				}
				if cmd.Flags().Changed("port") {
					formPort = strconv.FormatUint(uint64(portFlag), 10)
				}

				connValues, err := ui.RunConnectionForm(formDB, formUser, formHost, formPort)
				if err != nil {
					return err
				}

				finalDatabase = connValues.Database
				finalUser = connValues.Username
				finalHost = connValues.Host         // might be empty
				finalPassword = connValues.Password // might be empty
				if connValues.Port != "" {
					// ignoring error since the form validation ensures it's a valid port number
					finalPort = mustParsePort(connValues.Port)
				}
			} else {
				// non interactive mode, resolve database and user from flags and args
				finalDatabase, finalUser = resolveDBAndUser(string(dbNameFlag), string(usernameFlag), argDB, argUser)
				finalHost = string(hostFlag)
				finalPort = uint16(portFlag)
				if finalUser == "" {
					finalUser = getDefaultUser()
				}
			}

			postgres := database.New(cliCtx.Logger.Logger)
			cliCtx.Client = postgres

			var connector database.Connector
			var connErr error
			if strings.Contains(finalDatabase, "://") || strings.Contains(finalDatabase, "=") {
				connector, connErr = database.NewPGConnectorFromConnString(finalDatabase)
				if connErr != nil {
					cliCtx.Logger.Error("Invalid Connection string", "error", connErr)
					return connErr
				}

				cliCtx.Logger.Debug("Attempting database connection using connection string")
				connErr = cliCtx.Client.Connect(ctx, connector)
				if connErr != nil {
					cliCtx.Logger.Error("Failed to connect to database", "error", connErr)
					return connErr
				}
			} else {
				cliCtx.Logger.Debug("using field-based connection",
					"host", finalHost,
					"port", finalPort,
					"database", finalDatabase,
					"user", finalUser,
				)

				if bool(neverPromptFlag) && finalPassword == "" {
					finalPassword = getPasswordFromEnv()
				}

				if bool(forcePromptFlag) && finalPassword == "" {
					// Force prompt for password
					// TODO: Implement secure passowrd input
					pwd, promptErr := promptPassword()
					if promptErr != nil {
						return promptErr
					}
					finalPassword = pwd
				}

				connector, connErr = database.NewPGConnectorFromFields(
					finalHost,
					finalDatabase,
					finalUser,
					finalPassword,
					finalPort,
				)
				if connErr != nil {
					cliCtx.Logger.Error("Failed to create connector", "error", connErr)
					return connErr
				}

				cliCtx.Logger.Debug("Attempting database connection")
				connErr = cliCtx.Client.Connect(ctx, connector)
				if connErr != nil {
					if shouldAskForPassword(connErr, bool(neverPromptFlag)) {
						cliCtx.Logger.Debug("Connection failed, prompting for password")
						pwd, err := promptPassword()
						if err != nil {
							return err
						}
						connector.UpdatePassword(pwd)
						connRetryErr := cliCtx.Client.Connect(ctx, connector)
						if connRetryErr != nil {
							cliCtx.Logger.Error("Connection retry failed", "error", connRetryErr)
							return connRetryErr
						}
					} else {
						cliCtx.Logger.Error("Failed to connect to database", "error", connErr)
						return connErr
					}
				}
			}
			if !cliCtx.Client.IsConnected() {
				err := fmt.Errorf("failed to connect to database")
				cliCtx.Logger.Error("Failed to connect to database", "error", err)
				return err
			}

			app, err := app.NewPgxCLI(cliCtx.config, cliCtx.Printer, cliCtx.Client.Logger)
			if err != nil {
				cliCtx.Logger.Error("Failed to initialize app", "error", err)
				return err
			}
			cliCtx.App = app
			return nil
		},

		RunE: func(_ *cobra.Command, _ []string) error {
			if cliCtx.App == nil {
				cliCtx.Logger.Error("Application context not initialized")
				return fmt.Errorf("application context not initialized")
			}
			cliCtx.App.Start(ctx, cliCtx.Client)
			return nil
		},

		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if cliCtx.Logger != nil {
				if err := cliCtx.Logger.Close(); err != nil {
					return err
				}
			}
			if cliCtx.Client != nil {
				if err := cliCtx.Client.Close(ctx); err != nil {
					return err
				}
			}
			if cliCtx.App != nil {
				if err := cliCtx.App.Close(); err != nil {
					return err
				}
			}
			return nil
		},
	}

	// deactivating of the -h shorthand flag, so that it can be used in the host flag
	rootCmd.PersistentFlags().BoolP("help", "", false, "Print usage")
	_ = rootCmd.PersistentFlags().MarkShorthandDeprecated("help", "use --help")
	rootCmd.PersistentFlags().Lookup("help").Hidden = true

	debugFlag.bind(rootCmd)
	hostFlag.bind(rootCmd)
	portFlag.bind(rootCmd)
	dbNameFlag.bind(rootCmd)
	usernameFlag.bind(rootCmd)
	neverPromptFlag.bind(rootCmd)
	forcePromptFlag.bind(rootCmd)
	interactiveConnFlag.bind(rootCmd)

	rootCmd.MarkFlagsMutuallyExclusive("no-password", "password")

	return rootCmd
}

// getUserFromEnv gets username from environment variables
// support for pgcli specific environment variable
func getUserFromEnv() string {
	if userEnv := os.Getenv("PGXUSER"); userEnv != "" {
		return userEnv
	}
	if userEnv := os.Getenv("PGUSER"); userEnv != "" {
		return userEnv
	}
	return ""
}

// when database is given as flag then the next argument as user
func resolveDBAndUser(dbnameOpt, userOpt, argDB, argUser string) (string, string) {
	// Case: cmd -d database user
	if dbnameOpt != "" && argDB != "" && argUser == "" {
		return dbnameOpt, argDB
	}

	database := firstNonEmpty(dbnameOpt, argDB)
	user := firstNonEmpty(userOpt, argUser)
	return database, user
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func parsePositionalDBAndUser(args []string) (string, string) {
	var db string
	var user string
	if len(args) > 0 {
		db = args[0]
	}
	if len(args) > 1 {
		user = args[1]
	}
	return db, user
}

func promptPassword() (string, error) {
	var pwd string
	fmt.Print("Password: ")
	_, err := fmt.Scanln(&pwd)
	if err != nil {
		return "", err
	}
	return pwd, nil
}

func mustParsePort(port string) uint16 {
	portNum, _ := strconv.Atoi(port)
	return uint16(portNum)
}

// getDefaultUser gets the default username
// priority order:
// PGXUSER environment variable
// PGUSER environment variable
// OS user
func getDefaultUser() string {
	if user := getUserFromEnv(); user != "" {
		return user
	}
	return osUser
}

func getPasswordFromEnv() string {
	if passEnv := os.Getenv("PGXPASSWORD"); passEnv != "" {
		return passEnv
	}
	if passEnv := os.Getenv("PGPASSWORD"); passEnv != "" {
		return passEnv
	}
	return ""
}

func osUsername() string {
	currentUser, err := user.Current()
	if err != nil {
		return ""
	}
	username := currentUser.Username
	if strings.Contains(username, "\\") {
		username = username[strings.LastIndex(username, "\\")+1:]
	}
	return username
}
