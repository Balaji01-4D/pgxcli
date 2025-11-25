package cli

import (
	"context"
	"fmt"
	"os"
	"pgcli/internals/database"
	"pgcli/internals/repl"
	"strings"

	"github.com/spf13/cobra"
)


var (
	host string
	port uint16
	forcePrompt bool
	neverPrompt bool
	usernameOpt string
	dbnameOpt string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pgcli [DBNAME] [USERNAME]",
	Short: "Interactive PostgreSQL command-line client for querying and managing databases.",
	Version: GetVersion(),

	Args: cobra.MaximumNArgs(2),   // allowing maximum 2 args: DBNAME and USERNAME
	Run: func(cmd *cobra.Command, args []string) { 
		

		var argDB string		//  for storing positional DBNAME argument ex: pgcli mydb then argDB = "mydb"
		var argUser string 		// for storing positional USERNAME argument ex: pgcli mydb myuser then argUser = "myuser"

		if len(args) > 0 {
			argDB = args[0]   // first argument as DBNAME
		}
		if len(args) > 1 {
			argUser = args[1] // second argument as USERNAME
		}


		// when pgcli -d mydb myuser, here database name is given as flag then next arguement is considered as user
        db, _ := resolveDBAndUser(dbnameOpt, usernameOpt, argDB, argUser)
		// currently we dont use the user

		// currently supporting only DSN connection string example pgcli postgres://user:pass@localhost:5432/mydb
		if strings.Contains(db, "://") {
			fmt.Println("Connecting using DSN:", db)
			ctx := context.Background()

			// connecting to database using DSN
			// connection pool which manages multiple connections to the database
			pool, err := database.Connect(ctx, db)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
				os.Exit(1)
			}
			defer pool.Close()

			// creating executor to execute queries such as SELECT, INSERT, UPDATE, DELETE
			exec := database.NewExecutor(pool)
			// starting REPL with context and executor
			// because REPL needs to execute queries so passing executor
			// repl - read evaluate print loop
			repl.StartREPL(ctx, exec)

		}

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
