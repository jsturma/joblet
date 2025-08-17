package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	pb "joblet/api/gen"
	"joblet/internal/rnx/common"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	listWorkflow bool
)

// NewListCmd creates a new cobra command for listing jobs or workflows.
// The command supports JSON output format via the --json flag.
// Lists all jobs or workflows with their basic information.
func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all jobs or workflows",
		Long: `List all jobs or workflows in the system.

Examples:
  # List all jobs
  rnx list
  
  # List all workflows
  rnx list --workflow
  
  # List workflows in JSON format
  rnx list --workflow --json`,
		RunE: runList,
	}

	cmd.Flags().BoolVar(&listWorkflow, "workflow", false, "List workflows instead of jobs")

	return cmd
}

// runList executes the job or workflow listing command.
// Connects to the Joblet server, retrieves all jobs or workflows, and displays them
// in either human-readable table format or JSON format based on flags.
func runList(cmd *cobra.Command, args []string) error {
	if listWorkflow {
		return listWorkflows()
	}

	jobClient, err := common.NewJobClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer jobClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := jobClient.ListJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %v", err)
	}

	if len(response.Jobs) == 0 {
		if common.JSONOutput {
			fmt.Println("[]")
		} else {
			fmt.Println("No jobs found")
		}
		return nil
	}

	if common.JSONOutput {
		return outputJobsJSON(response.Jobs)
	}

	formatJobList(response.Jobs)

	return nil
}

func formatJobList(jobs []*pb.Job) {
	maxIDWidth := len("ID")
	maxStatusWidth := len("STATUS")

	// find the maximum width needed for each column
	for _, job := range jobs {
		if len(job.Id) > maxIDWidth {
			maxIDWidth = len(job.Id)
		}
		if len(job.Status) > maxStatusWidth {
			maxStatusWidth = len(job.Status)
		}
	}

	// some padding
	maxIDWidth += 2
	maxStatusWidth += 2

	// header
	fmt.Printf("%-*s %-*s %-19s %s\n",
		maxIDWidth, "ID",
		maxStatusWidth, "STATUS",
		"START TIME",
		"COMMAND")

	// separator line
	fmt.Printf("%s %s %s %s\n",
		strings.Repeat("-", maxIDWidth),
		strings.Repeat("-", maxStatusWidth),
		strings.Repeat("-", 19), // length of "START TIME"
		strings.Repeat("-", 7))  // length of "COMMAND"

	// each job
	for _, job := range jobs {

		startTime := formatStartTime(job.StartTime)

		// truncate long commands
		command := formatCommand(job.Command, job.Args)

		fmt.Printf("%-*s %-*s %-19s %s\n",
			maxIDWidth, job.Id,
			maxStatusWidth, job.Status,
			startTime,
			command)
	}
}

func formatStartTime(timeStr string) string {
	if timeStr == "" {
		return "N/A"
	}

	// Parse the RFC3339 timestamp
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return timeStr
	}

	return t.Format("2006-01-02 15:04:05")
}

func formatCommand(command string, args []string) string {
	if len(args) == 0 {
		return command
	}

	fullCommand := command + " " + strings.Join(args, " ")

	// truncate very long commands
	maxCommandLength := 80
	if len(fullCommand) > maxCommandLength {
		return fullCommand[:maxCommandLength-3] + "..."
	}

	return fullCommand
}

// outputJobsJSON outputs the jobs in JSON format
func outputJobsJSON(jobs []*pb.Job) error {
	// Convert protobuf jobs to a simpler structure for JSON output
	type jsonJob struct {
		ID            string   `json:"id"`
		Status        string   `json:"status"`
		StartTime     string   `json:"start_time"`
		EndTime       string   `json:"end_time,omitempty"`
		Command       string   `json:"command"`
		Args          []string `json:"args,omitempty"`
		ExitCode      int32    `json:"exit_code,omitempty"`
		MaxCPU        int32    `json:"max_cpu,omitempty"`
		MaxMemory     int32    `json:"max_memory,omitempty"`
		MaxIOBPS      int32    `json:"max_iobps,omitempty"`
		CPUCores      string   `json:"cpu_cores,omitempty"`
		ScheduledTime string   `json:"scheduled_time,omitempty"`
	}

	jsonJobs := make([]jsonJob, len(jobs))
	for i, job := range jobs {
		jsonJobs[i] = jsonJob{
			ID:            job.Id,
			Status:        job.Status,
			StartTime:     job.StartTime,
			EndTime:       job.EndTime,
			Command:       job.Command,
			Args:          job.Args,
			ExitCode:      job.ExitCode,
			MaxCPU:        job.MaxCPU,
			MaxMemory:     job.MaxMemory,
			MaxIOBPS:      job.MaxIOBPS,
			CPUCores:      job.CpuCores,
			ScheduledTime: job.ScheduledTime,
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonJobs)
}

// listWorkflows lists all workflows in the system
func listWorkflows() error {
	client, err := common.NewJobClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// Create workflow service client
	workflowClient := pb.NewJobServiceClient(client.GetConn())

	req := &pb.ListWorkflowsRequest{
		IncludeCompleted: true, // Always include all workflows
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := workflowClient.ListWorkflows(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to list workflows: %w", err)
	}

	if len(res.Workflows) == 0 {
		if common.JSONOutput {
			fmt.Println("[]")
		} else {
			fmt.Println("No workflows found")
		}
		return nil
	}

	if common.JSONOutput {
		return outputWorkflowsJSON(res.Workflows)
	}

	formatWorkflowList(res.Workflows)
	return nil
}

// formatWorkflowList formats and displays workflows in a table
func formatWorkflowList(workflows []*pb.WorkflowInfo) {
	fmt.Printf("ID   NAME                 STATUS      PROGRESS\n")
	fmt.Printf("---- -------------------- ----------- ---------\n")
	for _, workflow := range workflows {
		fmt.Printf("%-4d %-20s %-11s %d/%d\n",
			workflow.Id,
			truncateString(workflow.Name, 20),
			workflow.Status,
			workflow.CompletedJobs,
			workflow.TotalJobs)
	}
}

// outputWorkflowsJSON outputs the workflows in JSON format
func outputWorkflowsJSON(workflows []*pb.WorkflowInfo) error {
	// Convert protobuf workflows to a simpler structure for JSON output
	type jsonWorkflow struct {
		ID            int32  `json:"id"`
		Name          string `json:"name"`
		Template      string `json:"template"`
		Status        string `json:"status"`
		TotalJobs     int32  `json:"total_jobs"`
		CompletedJobs int32  `json:"completed_jobs"`
		FailedJobs    int32  `json:"failed_jobs"`
		CreatedAt     string `json:"created_at,omitempty"`
		StartedAt     string `json:"started_at,omitempty"`
		CompletedAt   string `json:"completed_at,omitempty"`
	}

	var jsonWorkflows []jsonWorkflow
	for _, workflow := range workflows {
		jsonWf := jsonWorkflow{
			ID:            workflow.Id,
			Name:          workflow.Name,
			Template:      workflow.Template,
			Status:        workflow.Status,
			TotalJobs:     workflow.TotalJobs,
			CompletedJobs: workflow.CompletedJobs,
			FailedJobs:    workflow.FailedJobs,
		}

		// Convert timestamps if present
		if workflow.CreatedAt != nil {
			jsonWf.CreatedAt = time.Unix(workflow.CreatedAt.Seconds, int64(workflow.CreatedAt.Nanos)).Format(time.RFC3339)
		}
		if workflow.StartedAt != nil {
			jsonWf.StartedAt = time.Unix(workflow.StartedAt.Seconds, int64(workflow.StartedAt.Nanos)).Format(time.RFC3339)
		}
		if workflow.CompletedAt != nil {
			jsonWf.CompletedAt = time.Unix(workflow.CompletedAt.Seconds, int64(workflow.CompletedAt.Nanos)).Format(time.RFC3339)
		}

		jsonWorkflows = append(jsonWorkflows, jsonWf)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonWorkflows)
}

func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}
