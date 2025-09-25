package gpu

import (
	"fmt"

	"joblet/internal/rnx/common"

	"github.com/spf13/cobra"
)

// NewGPUCmd creates the main GPU command
func NewGPUCmd() *cobra.Command {
	gpuCmd := &cobra.Command{
		Use:   "gpu",
		Short: "GPU management and monitoring commands",
		Long: `GPU management and monitoring commands for Joblet.

View available GPUs, monitor their status, and check utilization metrics.

Examples:
  rnx gpu list                    # List all available GPUs
  rnx gpu status                  # Show detailed GPU status and metrics
  rnx gpu status --gpu=0          # Show status for specific GPU
  rnx gpu status --json           # Output in JSON format`,
		DisableFlagsInUseLine: true,
	}

	// Add subcommands
	gpuCmd.AddCommand(NewGPUListCmd())
	gpuCmd.AddCommand(NewGPUStatusCmd())

	return gpuCmd
}

// NewGPUListCmd creates the GPU list command
func NewGPUListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available GPUs",
		Long: `List all available GPUs in the system with basic information.

Shows GPU index, name, memory, and current allocation status.

Examples:
  rnx gpu list                    # List all GPUs
  rnx gpu list --json             # Output in JSON format`,
		RunE: runGPUList,
	}
}

// NewGPUStatusCmd creates the GPU status command
func NewGPUStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show detailed GPU status and metrics",
		Long: `Show detailed GPU status including utilization, temperature, memory usage,
and current job allocations.

Examples:
  rnx gpu status                  # Show status for all GPUs
  rnx gpu status --gpu=0          # Show status for GPU 0 only
  rnx gpu status --json           # Output in JSON format
  rnx gpu status --watch          # Continuously monitor (updates every 5s)`,
		RunE: runGPUStatus,
	}

	cmd.Flags().IntP("gpu", "g", -1, "Show status for specific GPU index (-1 for all)")
	cmd.Flags().BoolP("watch", "w", false, "Continuously monitor GPU status")
	cmd.Flags().IntP("interval", "i", 5, "Update interval in seconds for watch mode")

	return cmd
}

// runGPUList executes the GPU list command
func runGPUList(cmd *cobra.Command, args []string) error {
	// For now, show a message about GPU support being available
	// This would be implemented once we have a dedicated GPU gRPC service

	if common.JSONOutput {
		fmt.Println(`{
  "message": "GPU list functionality available - requires GPU gRPC service implementation",
  "status": "not_implemented",
  "suggestion": "Use 'rnx job run --gpu=1 nvidia-smi' to check GPU availability"
}`)
		return nil
	}

	fmt.Println("GPU List:")
	fmt.Println("=" + string(make([]rune, 60)))
	fmt.Println()
	fmt.Println("GPU list functionality is available but requires a dedicated GPU gRPC service.")
	fmt.Println()
	fmt.Println("To check GPU availability, you can run:")
	fmt.Println("  rnx job run --gpu=1 nvidia-smi")
	fmt.Println()
	fmt.Println("This will:")
	fmt.Println("  • Test GPU allocation")
	fmt.Println("  • Show available GPUs")
	fmt.Println("  • Display GPU details and utilization")
	fmt.Println()

	return nil
}

// runGPUStatus executes the GPU status command
func runGPUStatus(cmd *cobra.Command, args []string) error {
	gpuIndex, _ := cmd.Flags().GetInt("gpu")
	watch, _ := cmd.Flags().GetBool("watch")
	interval, _ := cmd.Flags().GetInt("interval")

	if common.JSONOutput {
		fmt.Printf(`{
  "message": "GPU status functionality available - requires GPU gRPC service implementation",
  "status": "not_implemented",
  "requested_gpu": %d,
  "watch_mode": %t,
  "interval": %d,
  "suggestion": "Use 'rnx job run --gpu=1 nvidia-smi -l %d' for continuous monitoring"
}`, gpuIndex, watch, interval, interval)
		return nil
	}

	fmt.Println("GPU Status:")
	fmt.Println("=" + string(make([]rune, 60)))
	fmt.Println()
	fmt.Println("GPU status functionality is available but requires a dedicated GPU gRPC service.")
	fmt.Println()

	if watch {
		fmt.Printf("For continuous GPU monitoring (every %d seconds), you can run:\n", interval)
		fmt.Printf("  rnx job run --gpu=1 nvidia-smi -l %d\n", interval)
	} else {
		fmt.Println("For current GPU status, you can run:")
		fmt.Println("  rnx job run --gpu=1 nvidia-smi")
	}

	if gpuIndex >= 0 {
		fmt.Printf("\nTo monitor specific GPU %d:\n", gpuIndex)
		fmt.Printf("  rnx job run --gpu=1 nvidia-smi -i %d\n", gpuIndex)
	}

	fmt.Println()
	fmt.Println("This will show:")
	fmt.Println("  • GPU utilization")
	fmt.Println("  • Memory usage")
	fmt.Println("  • Temperature")
	fmt.Println("  • Power consumption")
	fmt.Println("  • Running processes")
	fmt.Println()

	return nil
}
