package cli

import (
	"github.com/spf13/cobra"
)

type Options struct {
	Host        string
	Port        uint16
	ForcePrompt bool
	NeverPrompt bool
	UsernameOpt string
	DBNameOpt   string
	Debug       bool
}

func bindFlags(cmd *cobra.Command, opts *Options) {
	cmd.Flags().StringVarP(&opts.Host, "host", "h", "", "host address of the postgres database")
	cmd.Flags().Uint16VarP(&opts.Port, "port", "p", 5432, "port number at which the postgres server is listening")
	cmd.Flags().StringVarP(&opts.UsernameOpt, "username", "u", "", "Username to connect to the postgres database.")
	cmd.Flags().StringVarP(&opts.UsernameOpt, "user", "U", "", "Username to connect to the postgres database.")

	cmd.Flags().BoolVarP(&opts.NeverPrompt, "no-password", "w", false, "never prompt for the password")
	cmd.Flags().BoolVarP(&opts.ForcePrompt, "password", "W", false, "Force password prompt")
	cmd.MarkFlagsMutuallyExclusive("password", "no-password")

	cmd.Flags().StringVarP(&opts.DBNameOpt, "dbname", "d", "", "database name to connect to.")
	cmd.Flags().BoolVar(&opts.Debug, "debug", false, "Enable debug mode for verbose logging.")
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
