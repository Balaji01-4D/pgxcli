package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	osuser "os/user"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/balaji01-4d/pgxcli/internal/config"
	"github.com/balaji01-4d/pgxcli/internal/database"
	"github.com/balaji01-4d/pgxcli/internal/logger"
	"github.com/balaji01-4d/pgxcli/internal/repl"
)

func run(_ *cobra.Command, args []string) {
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

	dbName, user := resolveDBAndUser(opts.DBNameOpt, opts.UsernameOpt, argDB, argUser)

	// Load config
	cfg := getConfig()

	logFilePath := cfg.Main.LogFile
	if logFilePath == "default" {
		configPath, err := config.GetConfigDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not get config dir for logging: %v\n", err)
		}
		logFilePath = filepath.Join(configPath, "pgxcli.log")
	}
	log := logger.InitLogger(opts.Debug, logFilePath)
	defer func() {
		if err := log.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close log file: %v\n", err)
		}
	}()

	// Log startup info
	log.Info("pgxcli starting",
		"version", version,
		"debug", opts.Debug,
	)
	log.Debug("parsed flags",
		"host", opts.Host,
		"port", opts.Port,
		"database", dbName,
		"user", user,
		"force_prompt", opts.ForcePrompt,
		"never_prompt", opts.NeverPrompt,
	)

	postgres := database.New(log.Logger)

	app := pgxCLI{
		config: cfg,
		client: postgres,
		repl:   repl.New(postgres, cfg, log.Logger),
		logger: log.Logger,
	}
	defer func() {
		if err := app.close(ctx); err != nil {
			log.Error("failed to close app", "error", err)
		}
	}()

	if user == "" {
		user = os.Getenv("PGUSER")
		if user == "" {
			currentUser, err := osuser.Current()
			if err != nil {
				log.Error("failed to get current user", "error", err)
				app.repl.PrintError(err)
				os.Exit(1)
			}
			user = currentUser.Username
			if strings.Contains(user, "\\") {
				user = user[strings.LastIndex(user, "\\")+1:]
			}
		}
	}
	if dbName == "" {
		dbName = user
	}

	log.Debug("resolved connection params",
		"database", dbName,
		"user", user,
		"connection_mode", getConnectionMode(dbName),
	)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		for {
			<-sigChan
		}
	}()

	appErr := app.start(ctx, dbName, user)
	if appErr != nil {
		log.Error("application error", "error", appErr)
		app.repl.PrintError(appErr)
	}
}

// getConnectionMode returns the connection mode based on the database string.
func getConnectionMode(db string) string {
	if strings.Contains(db, "://") || strings.Contains(db, "=") {
		return "connection_string"
	}
	return "fields"
}

func getConfig() *config.Config {
	cfg := config.DefaultConfig
	configDir, err := config.GetConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to get configuration directory, using the default configuration\n")
	}

	configPath, exists := config.CheckConfigExists(configDir)
	if exists {
		userCfg, err := config.LoadConfig(configPath)
		if err == nil {
			cfg = config.MergeConfig(cfg, userCfg)
		} else {
			fmt.Fprintf(os.Stderr, "unable to load user configuration\nerr:%v", err)
		}
	} else {
		err := config.SaveConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to save config %v\n", err)
		}
	}
	return &cfg
}
