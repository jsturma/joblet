package workflow

import (
	"github.com/spf13/cobra"
)

// NewWorkflowCmd creates the main workflow command
func NewWorkflowCmd() *cobra.Command {
	workflowCmd := &cobra.Command{
		Use:   "workflow",
		Short: "Manage workflows",
		Long: `Manage multi-job workflows with dependencies.

Workflows allow you to orchestrate multiple jobs with dependencies, creating
complex data pipelines, ML training workflows, and automated task sequences.

Examples:
  rnx workflow run pipeline.yaml           # Run a workflow
  rnx workflow list                        # List all workflows
  rnx workflow status <uuid>               # Check workflow status`,
		DisableFlagsInUseLine: true,
	}

	// Add subcommands
	workflowCmd.AddCommand(NewWorkflowRunCmd())
	workflowCmd.AddCommand(NewWorkflowListCmd())
	workflowCmd.AddCommand(NewWorkflowStatusCmd())

	return workflowCmd
}
