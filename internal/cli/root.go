package cli

import (
	"os"

	"github.com/spf13/cobra"
)

// opts holds all flag values bound to the root command.
var opts options

var rootCmd = &cobra.Command{
	Use:     "pgxcli [DBNAME] [USERNAME]",
	Short:   "Interactive PostgreSQL command-line client for querying and managing databases.",
	Version: version,
	Args:    cobra.MaximumNArgs(2), // Database name and username are optional example: pgxcli mydb myuser
	Run:     run,
}

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

	// bind all flags to opts
	bindFlags(rootCmd, &opts)
}
