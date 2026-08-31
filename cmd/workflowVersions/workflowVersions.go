package workflowVersions

import (
	"fmt"

	"github.com/IT-Academic-Research-Services/seqtoid-cli/pkg/seqtoid"
	"github.com/spf13/cobra"
)

// workflows are listed in the order shown to the user. The names match the workflow identifiers the
// server uses for the upload workflow_versions map and the catalog endpoint.
var workflows = []struct {
	name  string
	label string
}{
	{"short-read-mngs", "Metagenomics (Illumina, short-read-mngs)"},
	{"long-read-mngs", "Metagenomics (Nanopore, long-read-mngs)"},
	{"amr", "Antimicrobial Resistance (amr)"},
	{"consensus-genome", "Consensus Genome (consensus-genome)"},
}

// ManCmd lists the pipeline versions available for each workflow.
//
// Named "man" as requested; also reachable as "workflow-versions" or "versions", which describe
// what it does more plainly.
var ManCmd = &cobra.Command{
	Use:     "man",
	Aliases: []string{"workflow-versions", "versions"},
	Short:   "List the available pipeline versions for each workflow",
	Long: "List the pipeline versions you can select at upload with --workflow-version, grouped by " +
		"workflow. The newest version is marked (latest); deprecated versions still run but are no " +
		"longer patched. If a workflow shows no versions, per-run selection is not available in this " +
		"environment and uploads use the version configured for the project.",
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		for _, wf := range workflows {
			versions, err := seqtoid.DefaultClient.GetWorkflowVersions(wf.name)
			if err != nil {
				return fmt.Errorf("could not fetch versions for %s: %w", wf.name, err)
			}

			fmt.Fprintf(out, "%s\n", wf.label)
			if len(versions) == 0 {
				fmt.Fprintf(out, "  (no selectable versions published; uploads use the project's configured version)\n\n")
				continue
			}

			for i, v := range versions {
				tags := ""
				if i == 0 {
					tags += " (latest)"
				}
				if v.Deprecated {
					tags += " (deprecated)"
				}
				fmt.Fprintf(out, "  %s%s\n", v.Version, tags)
				if v.Notes != "" {
					fmt.Fprintf(out, "      %s\n", v.Notes)
				}
			}
			fmt.Fprintf(out, "\n")
		}
		return nil
	},
}
