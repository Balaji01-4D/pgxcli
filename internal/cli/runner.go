package cli

import (
	"context"
	"os"
	osuser "os/user"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/balaji01-4d/pgxcli/internal/config"
	"github.com/balaji01-4d/pgxcli/internal/database"
	"github.com/balaji01-4d/pgxcli/internal/logger"
	"github.com/balaji01-4d/pgxcli/internal/repl"
)

var (
	printErr = color.New(color.FgHiRed).FprintfFunc()
)

// run is the cobra Run func, extracted to keep root.go minimal.
func run(cmd *cobra.Command, args []string) {
	logger.InitLogger(opts.Debug, "logs/pgxcli.log")
	logger.Log.Info("pgxcli started")

	var argDB string
	var argUser string
	if len(args) > 0 {
		argDB = args[0]
	}
	if len(args) > 1 {
		argUser = args[1]
	}

	finalDB, finalUser := resolveDBAndUser(opts.DBNameOpt, opts.UsernameOpt, argDB, argUser)

	if finalUser == "" {
		currentser, err := osuser.Current()
		if err != nil {
			printErr(os.Stderr, "Failed to get current user: %v\n", err)
			os.Exit(1)
		}
		finalUser = currentser.Username
	}
	if finalDB == "" {
		finalDB = finalUser
	}

	// Load config
	cfg := config.DefaultConfig
	configDir, err := config.GetConfigDir()
	if err != nil {
		logger.Log.Error("Failed to get config directory", "error", err)
	} else {
		configPath, okay := config.CheckConfigExists(configDir)
		if !okay {
			logger.Log.Info("Config file does not exist, creating default config", "path", configPath)
			if err := config.SaveConfig(configPath, config.DefaultConfig); err != nil {
				logger.Log.Error("Failed to create default config file", "error", err)
			}
		} else {
			loadedConfig, err := config.LoadConfig(configPath)
			if err != nil {
				logger.Log.Error("Failed to load config file, using default config", "error", err)
			} else {
				cfg = loadedConfig
			}
		}
	}

	ctx := context.Background()

	postgres := database.New(opts.NeverPrompt, opts.ForcePrompt, ctx, cfg)
	replClient := repl.New(postgres, cfg)
	defer postgres.Close(ctx)

	var connector database.Connector

	if strings.Contains(finalDB, "://") || strings.Contains(finalDB, "=") {
		connector, err = database.NewPGConnectorFromConnString(finalDB)
		if err != nil {
			printErr(os.Stderr, "Invalid connection string: %v\n", err)
			os.Exit(1)
		}
	} else {
		var password string

		if opts.NeverPrompt {
			password = os.Getenv("PGPASSWORD")
		}

		if opts.ForcePrompt && password == "" {
			pwd, err := replClient.ReadPassword()
			if err != nil {
				printErr(os.Stderr, "Failed to read password: %v\n", err)
				os.Exit(1)
			}
			password = pwd
		}

		logger.Log.Debug("Connecting to database", "host", opts.Host, "port", opts.Port, "database", finalDB, "user", finalUser)
		connector, err = database.NewPGConnectorFromFields(
			opts.Host,
			finalDB,
			finalUser,
			password,
			opts.Port,
		)
		if err != nil {
			printErr(os.Stderr, "Failed to create connector: %v\n", err)
			os.Exit(1)
		}
	}

	ConnErr := postgres.Connect(ctx, connector)
	if ConnErr != nil {
		if shouldAskForPassword(ConnErr, opts.NeverPrompt) {
			pwd, err := replClient.ReadPassword()
			if err != nil {
				printErr(os.Stderr, "Failed to read password: %v\n", err)
				os.Exit(1)
			}
			connector.UpdatePassword(pwd)
			ConnErr = postgres.Connect(ctx, connector)
			if ConnErr != nil {
				printErr(os.Stderr, "%v\n", ConnErr)
				os.Exit(1)
			}
		} else {
			printErr(os.Stderr, "%v\n", ConnErr)
			os.Exit(1)
		}
	}

	if !postgres.IsConnected() {
		printErr(os.Stderr, "Not connected to any database\n")
		os.Exit(1)
	}

	replClient.Run(ctx)
	replClient.Close()
}
