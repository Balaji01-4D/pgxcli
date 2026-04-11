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
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to load configuration, using the default configuration\n")
		cfg, err = config.GetDefaultConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to get default configuration\n")
			os.Exit(1)
		}
	}

	log := logger.InitLogger(opts.Debug, cfg.Main.LogFile)
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
	repl, err := repl.New(postgres, cfg, log.Logger)
	if err != nil {
		log.Error("failed to initialize REPL", "error", err)
		fmt.Fprintf(os.Stderr, "failed to initialize REPL: %v\n", err)
		os.Exit(1)
	}

	app := pgxCLI{
		config: cfg,
		client: postgres,
		repl:   repl,
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