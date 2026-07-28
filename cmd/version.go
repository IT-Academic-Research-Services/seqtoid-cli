package cmd

import (
	"fmt"

	"github.com/IT-Academic-Research-Services/seqtoid-cli/pkg"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the current version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(pkg.Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
