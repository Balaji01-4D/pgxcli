package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	osuser "os/user"
	"strings"

	"github.com/spf13/cobra"

	"github.com/balaji01-4d/pgxcli/internal/config"
	"github.com/balaji01-4d/pgxcli/internal/database"
	"github.com/balaji01-4d/pgxcli/internal/logger"
	"github.com/balaji01-4d/pgxcli/internal/repl"
)

func run(cmd *cobra.Command, args []string) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	
	var argDB string
	var argUser string
	if len(args) > 0 {
		argDB = args[0]
	}
	if len(args) > 1 {
		argUser = args[1]
	}
	
	finalDB, finalUser := resolveDBAndUser(opts.DBNameOpt, opts.UsernameOpt, argDB, argUser)

	// Load config
	cfg := getConfig()

	logger.InitLogger(opts.Debug, "logs/pgxcli.log")

	postgres := database.New(ctx, cfg)
	
	r := repl.New(postgres, cfg)

	if finalUser == "" {
		finalUser = os.Getenv("PGUSER")
		if finalUser == "" {
			currentUser, err := osuser.Current()
			if err != nil {
				r.PrintError(err)
				os.Exit(1)
			}
			finalUser = currentUser.Username
		}
	}
	if finalDB == "" {
		finalDB = finalUser
	}
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	
	go func() {
		for {
			<-sigChan
		}
	}()

	appErr := startApplication(ctx, cfg, finalDB, finalUser)
	if appErr != nil {
		r.PrintError(appErr)
	}
}

func startApplication(ctx context.Context, cfg config.Config, db, user string) error {
	postgres := database.New(ctx, cfg)
	defer postgres.Close(ctx)

	replClient := repl.New(postgres, cfg)
	defer replClient.Close()

	var connector database.Connector
	var err error

	if strings.Contains(db, "://") || strings.Contains(db, "=") {
		connector, err = database.NewPGConnectorFromConnString(db)
		if err != nil {
			return fmt.Errorf("Invalid connection string: %v\n", err)
		}
	} else {
		var password string

		if opts.NeverPrompt {
			password = os.Getenv("PGPASSWORD")
		}

		if opts.ForcePrompt && password == "" {
			pwd, err := replClient.ReadPassword()
			if err != nil {
				return fmt.Errorf("Failed to read password: %v\n", err)
			}
			password = pwd
		}

		logger.Log.Debug("Connecting to database", "host", opts.Host, "port", opts.Port, "database", db, "user", user)
		connector, err = database.NewPGConnectorFromFields(
			opts.Host,
			db,
			user,
			password,
			opts.Port,
		)
		if err != nil {
			return fmt.Errorf("Failed to create connector: %v\n", err)
		}
	}

	ConnErr := postgres.Connect(ctx, connector)
	if ConnErr != nil {
		if shouldAskForPassword(ConnErr, opts.NeverPrompt) {
			pwd, err := replClient.ReadPassword()
			if err != nil {
				return fmt.Errorf("Failed to read password: %v\n", err)
			}
			connector.UpdatePassword(pwd)
			ConnErr = postgres.Connect(ctx, connector)
			if ConnErr != nil {
				return err
			}
		} else {
			return ConnErr
		}
	}

	if !postgres.IsConnected() {
		return fmt.Errorf("Not connected to any database\n")
	}

	replClient.Run(ctx)
	replClient.Close()

	return nil
}


func getConfig() config.Config {
	cfg := config.DefaultConfig
	configDir, err := config.GetConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to get configuartion directory, using the default configuration\n")
	}

	configPath, exists := config.CheckConfigExists(configDir)
	if exists {
		userCfg, err := config.LoadConfig(configPath)
		if err == nil {
			fmt.Fprintf(os.Stderr, "unable to load user configuration\nerr:%v", err)
			cfg = userCfg
		}
	} else {
		err := config.SaveConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to save config %v\n", err)
		}
	}
	return cfg
}