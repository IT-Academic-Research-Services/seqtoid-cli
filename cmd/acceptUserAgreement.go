package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var autoAccept bool

// acceptUserAgreementCmd represents the acceptUserAgreement command
var acceptUserAgreementCmd = &cobra.Command{
	Use:   "accept-user-agreement",
	Short: "Accept the SeqToID User Agreement",
	Long:  "Accept the SeqToID User Agreement",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print(`I agree that the data I am uploading to SeqToID
has been lawfully collected and that I have all the necessary consents,
permissions, and authorizations needed to collect, share, and export data to
SeqToID as outlined in the Terms (https://seqtoid.org/terms) and Data
Privacy Notice (https://seqtoid.org/privacy).

Accept (y/n)? y/Y for yes (any other input to cancel): `)
		accepted := false
		if !autoAccept {
			input := bufio.NewScanner(os.Stdin)
			input.Scan()
			response := input.Text()
			accepted = strings.ToLower(response) == "y"
		}
		if accepted || autoAccept {
			viper.Set("accepted_user_agreement", "Y")
			return viper.WriteConfig()
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(acceptUserAgreementCmd)
	acceptUserAgreementCmd.Flags().BoolP("yes", "y", false, "Accept without prompting")
}
