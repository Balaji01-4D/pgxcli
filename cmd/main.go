/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"pgcli/internals/pg"
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


func main() {
	Execute()
}


// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pgcli [DBNAME] [USERNAME]",
	Short: "Interactive PostgreSQL command-line client for querying and managing databases.",
	Long: `pgcli is an interactive PostgreSQL command-line client for connecting to databases, running SQL queries, and inspecting schema objects.
It aims to provide a simple, scriptable interface for everyday database tasks such as querying, debugging, and administration.`,
	Version: "0.0.1",

	Args: cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) { 
		

		var argDB string
		var argUser string 

		if len(args) > 0 {
			argDB = args[0]
		}
		if len(args) > 1 {
			argUser = args[1]
		}


        database, user := resolveDBAndUser(dbnameOpt, usernameOpt, argDB, argUser)

        fmt.Println("Final Database:", database)
        fmt.Println("Final User:", user)


		if strings.Contains(database, "://") {
			ctx := context.Background()
			pool, err := pg.Connect(ctx, database)
			if err != nil {
				panic(err)
			}
			defer pool.Close()

			exec := pg.NewExecutor(pool)
			st, err := exec.Query(ctx, "SELECT * FROM students")
			if err != nil {
				panic(err)
			}
			defer st.Close()
			fmt.Println(st.Columns())
			fmt.Println(st.Duration())

			for {
				row, err := st.Next()
				if err != nil {
					if err == io.EOF {
						break
					} else {
						panic(err)
					}
				}
				fmt.Println(row)
			}
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
