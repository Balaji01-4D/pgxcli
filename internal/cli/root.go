package cli

import (
	"context"
	"os"
	"strings"

	"github.com/balaji01-4d/pgxcli/internal/config"
	"github.com/balaji01-4d/pgxcli/internal/database"
	"github.com/balaji01-4d/pgxcli/internal/logger"
	"github.com/balaji01-4d/pgxcli/internal/repl"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	host        string
	port        uint16
	forcePrompt bool
	neverPrompt bool
	usernameOpt string
	dbnameOpt   string
	debug       bool
)

var (
	printErr  = color.New(color.FgHiRed).FprintfFunc()
	printTime = color.New(color.FgHiCyan).FprintfFunc()
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "pgxcli [DBNAME] [USERNAME]",
	Short:   "Interactive PostgreSQL command-line client for querying and managing databases.",
	Version: GetVersion(),

	Args: cobra.MaximumNArgs(2), // allowing maximum 2 args: DBNAME and USERNAME
	Run: func(cmd *cobra.Command, args []string) {

		logger.InitLogger(debug, "logs/pgxcli.log")
		logger.Log.Info("pgxcli started")

		var argDB string   //  for storing positional DBNAME argument ex: pgxcli mydb then argDB = "mydb"
		var argUser string // for storing positional USERNAME argument ex: pgxcli mydb myuser then argUser = "myuser"

		if len(args) > 0 {
			argDB = args[0] // first argument as DBNAME
		}
		if len(args) > 1 {
			argUser = args[1] // second argument as USERNAME
		}

		// when pgxcli -d mydb myuser, here database name is given as flag then next arguement is considered as user
		finalDB, finalUser := resolveDBAndUser(dbnameOpt, usernameOpt, argDB, argUser)
		// currently we dont use the user

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

		postgres := database.New(neverPrompt, forcePrompt, ctx, cfg)
		defer postgres.Close(ctx)

		if strings.Contains(finalDB, "://") {
			err := postgres.ConnectURI(ctx, finalDB)
			if err != nil {
				printErr(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		} else if strings.Contains(finalDB, "=") {
			err := postgres.ConnectDSN(ctx, finalDB)
			if err != nil {
				printErr(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		} else {
			logger.Log.Info("Connecting to database", "host", host, "port", port, "database", finalDB, "user", finalUser)
			err := postgres.Connect(ctx, host, finalUser, "", finalDB, "", port)
			if err != nil {
				logger.Log.Error("Connection failed", "error", err, "host", host, "database", finalDB)
				printErr(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		}
		if !postgres.IsConnected() {
			printErr(os.Stderr, "Not connected to any database\n")
			os.Exit(1)
		}

		repl := repl.New(postgres, cfg)
		repl.Run(ctx)
		repl.Close()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	// deactivating of the -h shorthand flag, so that it can be used in the host flag
	rootCmd.PersistentFlags().BoolP("help", "", false, "Print usage")
	rootCmd.PersistentFlags().MarkShorthandDeprecated("help", "use --help")
	rootCmd.PersistentFlags().Lookup("help").Hidden = true

	rootCmd.Flags().StringVarP(&host, "host", "h", "", "host address of the postgres database")
	rootCmd.Flags().Uint16VarP(&port, "port", "p", 5432, "port number at which the postgres server is listening")
	rootCmd.Flags().StringVarP(&usernameOpt, "username", "u", "", "Username to connect to the postgres database.")
	rootCmd.Flags().StringVarP(&usernameOpt, "user", "U", "", "Username to connect to the postgres database.")

	rootCmd.Flags().BoolVarP(&neverPrompt, "no-password", "w", false, "never prompt for the password")
	rootCmd.Flags().BoolVarP(&forcePrompt, "password", "W", false, "Force password prompt")
	rootCmd.MarkFlagsMutuallyExclusive("password", "no-password")

	rootCmd.Flags().StringVarP(&dbnameOpt, "dbname", "d", "", "database name to connect to.")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode for verbose logging.")

}

// when database is given as flag then the next argument as user
func resolveDBAndUser(dbnameOpt, userOpt, argDB, argUser string) (string, string) {

	// Case:cmd -d database user
	if dbnameOpt != "" && argDB != "" && argUser == "" {
		return dbnameOpt, argDB
	}

	// Normal resolution priority
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
