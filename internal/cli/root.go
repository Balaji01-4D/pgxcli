package cli

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"

	"golang.org/x/term"

	"github.com/balaji01-4d/pgxcli/internal/app"
	"github.com/balaji01-4d/pgxcli/internal/completer"
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

// NewRootCmd builds the root cobra command and wires the CLI lifecycle hooks.
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

	rootCmd := &cobra.Command{
		Use:     "pgxcli [DBNAME] [USERNAME]",
		Short:   "Interactive PostgreSQL command-line client for querying and managing databases.",
		Version: version,
		Args:    cobra.MaximumNArgs(2), // Database name and username are optional example: pgxcli mydb myuser

		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return loadRuntimeDependencies(cliCtx, bool(debugFlag))
		},

		PreRunE: func(cmd *cobra.Command, args []string) error {
			params, err := resolveConnectionParams(
				cmd,
				args,
				bool(interactiveConnFlag),
				string(dbNameFlag),
				string(usernameFlag),
				string(hostFlag),
				uint16(portFlag),
			)
			if err != nil {
				return err
			}
			if err := connectClient(ctx, cliCtx, params, bool(neverPromptFlag), bool(forcePromptFlag)); err != nil {
				return err
			}
			if err := ensureConnected(cliCtx); err != nil {
				return err
			}
			return initApplication(cliCtx)
		},

		RunE: func(_ *cobra.Command, _ []string) error {
			if cliCtx.App == nil {
				cliCtx.Logger.Error("Application context not initialized")
				return fmt.Errorf("application context not initialized")
			}
			if !bool(interactiveConnFlag) {
				ui.PrintBanner(version)
			}
			return cliCtx.App.Start(ctx, cliCtx.Client)
		},

		PersistentPostRunE: func(_ *cobra.Command, _ []string) error {
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

type connectionParams struct {
	database string
	user     string
	host     string
	port     uint16
	password string
}

// loadRuntimeDependencies loads configuration, initializes logger, and sets up pager.
func loadRuntimeDependencies(cliCtx *CliContext, debug bool) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cliCtx.config = cfg

	if err := cliCtx.Printer.SetPagerMode(cfg.Main.Pager); err != nil { // Printer is initialized in main.go.
		return err
	}

	initializedLogger, err := logger.InitLogger(debug, cfg.Main.LogFile)
	if err != nil {
		return err
	}
	cliCtx.Logger = initializedLogger

	return nil
}

func resolveConnectionParams(
	cmd *cobra.Command,
	args []string,
	interactive bool,
	dbnameOpt string,
	userOpt string,
	hostOpt string,
	portOpt uint16,
) (connectionParams, error) {
	argDB, argUser := parsePositionalDBAndUser(args)
	if interactive {
		return resolveInteractiveConnectionParams(cmd, argDB, argUser, dbnameOpt, userOpt, hostOpt, portOpt)
	}

	database, user := resolveDBAndUser(dbnameOpt, userOpt, argDB, argUser)
	if user == "" {
		user = getDefaultUser()
	}

	return connectionParams{
		database: database,
		user:     user,
		host:     hostOpt,
		port:     portOpt,
	}, nil
}

func resolveInteractiveConnectionParams(
	cmd *cobra.Command,
	argDB string,
	argUser string,
	dbnameOpt string,
	userOpt string,
	hostOpt string,
	portOpt uint16,
) (connectionParams, error) {
	// In interactive mode, flags / args are used as defaults.
	// Priority is flag, arg, env, default.
	formDB := firstNonEmpty(dbnameOpt, argDB)
	formUser := firstNonEmpty(userOpt, argUser, getDefaultUser())

	var formHost string
	var formPort string
	if cmd.Flags().Changed("host") {
		formHost = hostOpt
	}
	if cmd.Flags().Changed("port") {
		formPort = strconv.FormatUint(uint64(portOpt), 10)
	}

	connValues, err := ui.RunConnectionForm(formDB, formUser, formHost, formPort)
	if err != nil {
		return connectionParams{}, err
	}

	params := connectionParams{
		database: connValues.Database,
		user:     connValues.Username,
		host:     connValues.Host,
		password: connValues.Password,
	}
	if connValues.Port != "" {
		// Ignoring error since the form validation ensures this is a valid port.
		params.port, err = mustParsePort(connValues.Port)
		if err != nil {
			return connectionParams{}, err
		}
	}

	return params, nil
}

func connectClient(
	ctx context.Context,
	cliCtx *CliContext,
	params connectionParams,
	neverPrompt bool,
	forcePrompt bool,
) error {
	cliCtx.Client = database.New(cliCtx.Logger.Logger)

	if strings.Contains(params.database, "://") || strings.Contains(params.database, "=") {
		return connectWithConnString(ctx, cliCtx, params.database)
	}

	return connectWithFields(ctx, cliCtx, params, neverPrompt, forcePrompt)
}

func connectWithConnString(ctx context.Context, cliCtx *CliContext, connString string) error {
	connector, err := database.NewPGConnectorFromConnString(connString)
	if err != nil {
		cliCtx.Logger.Error("Invalid Connection string", "error", err)
		return err
	}

	cliCtx.Logger.Debug("Attempting database connection using connection string")
	if err := cliCtx.Client.Connect(ctx, connector); err != nil {
		cliCtx.Logger.Error("Failed to connect to database", "error", err)
		return err
	}

	return nil
}

func connectWithFields(
	ctx context.Context,
	cliCtx *CliContext,
	params connectionParams,
	neverPrompt bool,
	forcePrompt bool,
) error {
	password := params.password

	cliCtx.Logger.Debug("using field-based connection",
		"host", params.host,
		"port", params.port,
		"database", params.database,
		"user", params.user,
	)

	if neverPrompt && password == "" {
		password = getPasswordFromEnv()
	}

	if forcePrompt && password == "" {
		pwd, err := promptPassword("Enter password")
		if err != nil {
			return err
		}
		password = pwd
	}

	connector, err := database.NewPGConnectorFromFields(
		params.host,
		params.database,
		params.user,
		password,
		params.port,
	)
	if err != nil {
		cliCtx.Logger.Error("Failed to create connector", "error", err)
		return err
	}

	cliCtx.Logger.Debug("Attempting database connection")
	connErr := cliCtx.Client.Connect(ctx, connector)
	if connErr == nil {
		return nil
	}

	if !shouldAskForPassword(connErr, neverPrompt) {
		cliCtx.Logger.Error("Failed to connect to database", "error", connErr)
		return connErr
	}

	cliCtx.Logger.Debug("Connection failed, prompting for password")
	fmt.Fprintln(os.Stderr, "Wrong password, try again.")
	pwd, err := promptPassword("Enter password again")
	if err != nil {
		return err
	}

	connector.UpdatePassword(pwd)
	if connRetryErr := cliCtx.Client.Connect(ctx, connector); connRetryErr != nil {
		cliCtx.Logger.Error("Connection retry failed", "error", connRetryErr)
		return connRetryErr
	}

	return nil
}

func ensureConnected(cliCtx *CliContext) error {
	if cliCtx.Client.IsConnected() {
		return nil
	}

	err := fmt.Errorf("failed to connect to database")
	cliCtx.Logger.Error("Failed to connect to database", "error", err)
	return err
}

// initApplication Initializes the app,
// which includes setting up the logger, config and autocompleter with PostgreSQL keywords.
func initApplication(cliCtx *CliContext) error {
	completer := completer.New(cliCtx.Logger.Logger)
	pgxCLI, err := app.New(cliCtx.config, cliCtx.Printer, cliCtx.Logger.Logger, completer)
	if err != nil {
		cliCtx.Logger.Error("Failed to initialize app", "error", err)
		return err
	}

	cliCtx.App = pgxCLI
	return nil
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

func promptPassword(s string) (string, error) {
	fmt.Printf("%s: ", s)
	fd := int(os.Stdin.Fd())
	oldState, err := term.GetState(fd)
	if err != nil {
		// stdin is not a TTY — fall back to normal line input
		var pwd string
		_, err := fmt.Scanln(&pwd)
		if err != nil {
			return "", err
		}
		return pwd, nil
	}

	// Put terminal in raw mode (echo off) for secure password entry.
	if _, err := term.MakeRaw(fd); err != nil {
		return "", fmt.Errorf("failed to set raw terminal mode: %w", err)
	}

	pwd, err := term.ReadPassword(fd)
	// Restore terminal to its original state no matter what.
	_ = term.Restore(fd, oldState)
	fmt.Println()

	if err != nil {
		return "", err
	}
	return string(pwd), nil
}

func mustParsePort(port string) (uint16, error) {
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return 0, err
	}
	return uint16(portNum), nil
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
