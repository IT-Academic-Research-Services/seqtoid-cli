package amr

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/IT-Academic-Research-Services/seqtoid-cli/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var projectName string
var rawMetadata []string
var stringMetadata map[string]string
var metadataCSVPath string
var disableBuffer bool
var workflowVersion string

// AmrCmd represents the Amr command
var AmrCmd = &cobra.Command{
	Use:   "amr",
	Short: "Commands related to the amr pipeline",
	Long:  "Commands related to the amr pipeline",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if strings.ToLower(viper.GetString("accepted_user_agreement")) != "y" {
			fmt.Println("Cannot upload samples until the user agreement is accepted, run seqtoid accept-user-agreement or set SEQTOID_CLI_ACCEPTED_USER_AGREEMENT=Y")
			os.Exit(2)
		}
	},
}

func loadSharedFlags(c *cobra.Command) {
	c.Flags().StringVarP(&projectName, "project", "p", "", "Project name. Make sure the project is created on the website (required)")
	c.Flags().StringArrayVarP(&rawMetadata, "metadatum", "m", nil, "Metadatum name and value for your sample, ex. 'host=Human'. Repeat -m for multiple; values may contain commas (e.g. a location).")
	c.Flags().StringVar(&metadataCSVPath, "metadata-csv", "", "Metadata local file path.")
	c.Flags().BoolVar(&disableBuffer, "disable-buffer", false, "Disable shared buffer pool (useful if running out of memory)")
	c.Flags().StringVar(&workflowVersion, "workflow-version", "", "Pipeline version to run for this upload, e.g. '1.4.0'. Optional; defaults to the version configured for the project. Run 'seqtoid man' to list available versions.")
}

func validateCommonArgs() error {
	parsed, err := util.ParseMetadataPairs(rawMetadata)
	if err != nil {
		return err
	}
	stringMetadata = parsed

	if projectName == "" {
		return errors.New("missing required argument: project")
	}

	return nil
}
