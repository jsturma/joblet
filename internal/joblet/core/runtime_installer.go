//go:build linux

package core

import (
	"archive/zip"
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"joblet/pkg/config"
	"joblet/pkg/logger"
	"joblet/pkg/platform"
)

// RuntimeInstallationStreamer interface for streaming runtime installation progress
type RuntimeInstallationStreamer interface {
	SendProgress(message string) error
	SendLog(data []byte) error
}

// RuntimeInstaller handles runtime installation in a dedicated chroot environment
// This is separate from the job execution system and provides its own isolation
type RuntimeInstaller struct {
	config   *config.Config
	logger   *logger.Logger
	platform platform.Platform
}

// NewRuntimeInstaller creates a new runtime installer
func NewRuntimeInstaller(config *config.Config, logger *logger.Logger, platform platform.Platform) *RuntimeInstaller {
	return &RuntimeInstaller{
		config:   config,
		logger:   logger.WithField("component", "runtime-installer"),
		platform: platform,
	}
}

// buildPathFromConfig creates a PATH environment variable from configured common paths
func (ri *RuntimeInstaller) buildPathFromConfig() string {
	pathParts := make([]string, 0, len(ri.config.Runtime.CommonPaths)+4)

	// Add common paths from config
	pathParts = append(pathParts, ri.config.Runtime.CommonPaths...)

	// Add essential system paths not in common paths
	systemPaths := []string{"/bin", "/sbin", "/usr/sbin"}
	for _, sysPath := range systemPaths {
		found := false
		for _, commonPath := range ri.config.Runtime.CommonPaths {
			if commonPath == sysPath {
				found = true
				break
			}
		}
		if !found {
			pathParts = append(pathParts, sysPath)
		}
	}

	return "PATH=" + strings.Join(pathParts, ":")
}

// RuntimeInstallRequest represents a runtime installation request
type RuntimeInstallRequest struct {
	RuntimeSpec    string
	Repository     string
	Branch         string
	Path           string
	ForceReinstall bool
	Streamer       RuntimeInstallationStreamer // Optional streaming callback
}

// RuntimeInstallResult represents the result of a runtime installation
type RuntimeInstallResult struct {
	RuntimeSpec string
	Success     bool
	Message     string
	InstallPath string
	Duration    time.Duration
	LogOutput   string
}

// InstallFromGithub installs a runtime from local development workspace first,
// then falls back to GitHub repository if local installation fails
func (ri *RuntimeInstaller) InstallFromGithub(ctx context.Context, req *RuntimeInstallRequest) (*RuntimeInstallResult, error) {

	startTime := time.Now()

	ri.logger.Info("installing runtime from GitHub", "spec", req.RuntimeSpec, "repo", req.Repository, "branch", req.Branch, "force", req.ForceReinstall)

	// Handle force reinstall: cleanup existing runtime if it exists
	if req.ForceReinstall {
		runtimePath := ri.getRuntimePath(req.RuntimeSpec)
		fullRuntimePath := filepath.Join(ri.config.Runtime.BasePath, runtimePath)

		if _, err := ri.platform.Stat(fullRuntimePath); err == nil {
			ri.logger.Info("force reinstall requested, removing existing runtime", "path", fullRuntimePath)
			if err := ri.platform.RemoveAll(fullRuntimePath); err != nil {
				ri.logger.Warn("failed to remove existing runtime", "path", fullRuntimePath, "error", err)
				// Continue anyway - let the installation proceed
			} else {
				ri.logger.Info("existing runtime removed successfully", "path", fullRuntimePath)
			}
		}
	}

	// Skip local installation for remote servers - use GitHub directly
	// Local runtime detection should be handled by the client

	// Set defaults for GitHub
	repository := req.Repository
	if repository == "" {
		repository = "ehsaniara/joblet"
	}

	branch := req.Branch
	if branch == "" {
		branch = "main"
	}

	resolvedPath := req.Path
	if resolvedPath == "" {
		resolvedPath = ri.autoDetectRuntimePath(req.RuntimeSpec)
	}

	// Create dedicated builder chroot environment with full host OS access
	chrootDir, cleanup, err := ri.createBuilderChroot()
	if err != nil {
		return nil, fmt.Errorf("failed to create builder chroot: %w", err)
	}
	defer cleanup()

	// Execute installation in chroot with optional streaming
	result, err := ri.executeInstallationInChrootWithStreaming(ctx, chrootDir, repository, branch, resolvedPath, req.RuntimeSpec, req.Streamer)
	if err != nil {
		return &RuntimeInstallResult{
			RuntimeSpec: req.RuntimeSpec,
			Success:     false,
			Message:     fmt.Sprintf("GitHub installation failed: %v", err),
			Duration:    time.Since(startTime),
		}, err
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// RuntimeInstallFromLocalRequest represents a request to install runtime from uploaded files
type RuntimeInstallFromLocalRequest struct {
	RuntimeSpec    string
	Files          []*RuntimeFile
	ForceReinstall bool
	Streamer       RuntimeInstallationStreamer // Optional streaming callback
}

// RuntimeFile represents an uploaded runtime file
type RuntimeFile struct {
	Path       string
	Content    []byte
	Executable bool
}

// InstallFromLocal installs a runtime from uploaded local development files
func (ri *RuntimeInstaller) InstallFromLocal(ctx context.Context, req *RuntimeInstallFromLocalRequest) (*RuntimeInstallResult, error) {
	startTime := time.Now()

	ri.logger.Info("installing runtime from local uploaded files", "spec", req.RuntimeSpec, "files", len(req.Files), "force", req.ForceReinstall)

	// Handle force reinstall: cleanup existing runtime if it exists
	if req.ForceReinstall {
		runtimePath := ri.getRuntimePath(req.RuntimeSpec)
		fullRuntimePath := filepath.Join(ri.config.Runtime.BasePath, runtimePath)

		if _, err := ri.platform.Stat(fullRuntimePath); err == nil {
			ri.logger.Info("force reinstall requested, removing existing runtime", "path", fullRuntimePath)
			if err := ri.platform.RemoveAll(fullRuntimePath); err != nil {
				ri.logger.Warn("failed to remove existing runtime", "path", fullRuntimePath, "error", err)
				// Continue anyway - let the installation proceed
			} else {
				ri.logger.Info("existing runtime removed successfully", "path", fullRuntimePath)
			}
		}
	}

	// Create dedicated builder chroot environment with full host OS access
	chrootDir, cleanup, err := ri.createBuilderChroot()
	if err != nil {
		return nil, fmt.Errorf("failed to create builder chroot: %w", err)
	}
	defer cleanup()

	// Execute installation in chroot with uploaded files and optional streaming
	result, err := ri.executeLocalInstallationInChrootWithStreaming(ctx, chrootDir, req.RuntimeSpec, req.Files, req.Streamer)
	if err != nil {
		return &RuntimeInstallResult{
			RuntimeSpec: req.RuntimeSpec,
			Success:     false,
			Message:     fmt.Sprintf("Local installation failed: %v", err),
			Duration:    time.Since(startTime),
		}, err
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// executeLocalInstallationInChrootWithStreaming executes local installation with streaming support
func (ri *RuntimeInstaller) executeLocalInstallationInChrootWithStreaming(ctx context.Context, chrootDir, runtimeSpec string, files []*RuntimeFile, streamer RuntimeInstallationStreamer) (*RuntimeInstallResult, error) {
	log := ri.logger.WithField("chrootDir", chrootDir)
	log.Info("executing local runtime installation in chroot with streaming")

	// Send initial progress
	if streamer != nil {
		if err := streamer.SendProgress("🔧 Starting local runtime installation..."); err != nil {
			log.Warn("failed to send progress", "error", err)
		}
	}

	// Step 1: Write uploaded files to /tmp in the chroot (not host /tmp)
	if streamer != nil {
		if err := streamer.SendProgress("📁 Writing runtime files to chroot..."); err != nil {
			log.Warn("failed to send progress", "error", err)
		}
	}

	runtimeScriptsDir := "/tmp/runtime-scripts"
	if err := ri.writeFilesToChroot(chrootDir, runtimeScriptsDir, files); err != nil {
		return &RuntimeInstallResult{
			RuntimeSpec: runtimeSpec,
			Success:     false,
			Message:     fmt.Sprintf("Failed to write runtime files: %v", err),
		}, err
	}

	// Step 2: Execute setup script in chroot with streaming
	if streamer != nil {
		if err := streamer.SendProgress("🏗️  Running local setup script..."); err != nil {
			log.Warn("failed to send progress", "error", err)
		}
	}

	output, err := ri.executeLocalSetupInChrootWithStreaming(ctx, chrootDir, runtimeScriptsDir, runtimeSpec, streamer)
	if err != nil {
		return &RuntimeInstallResult{
			RuntimeSpec: runtimeSpec,
			Success:     false,
			Message:     fmt.Sprintf("Setup execution failed: %v", err),
			LogOutput:   output,
		}, err
	}

	// Step 3: Copy runtime from chroot to host
	if streamer != nil {
		if err := streamer.SendProgress("📦 Copying runtime to host..."); err != nil {
			log.Warn("failed to send progress", "error", err)
		}
	}

	hostRuntimePath := filepath.Join(ri.config.Runtime.BasePath, ri.getRuntimePath(runtimeSpec))
	chrootRuntimePath := filepath.Join(chrootDir, "opt/joblet/runtimes", ri.getRuntimePath(runtimeSpec))

	// Create host runtime directory
	if err := ri.platform.MkdirAll(hostRuntimePath, 0755); err != nil {
		return &RuntimeInstallResult{
			RuntimeSpec: runtimeSpec,
			Success:     false,
			Message:     fmt.Sprintf("Failed to create host runtime directory: %v", err),
			LogOutput:   output,
		}, err
	}

	// Copy from chroot to host
	copyCmd := exec.CommandContext(ctx, "cp", "-r", chrootRuntimePath+"/.", hostRuntimePath)
	if err := copyCmd.Run(); err != nil {
		return &RuntimeInstallResult{
			RuntimeSpec: runtimeSpec,
			Success:     false,
			Message:     fmt.Sprintf("Failed to copy runtime from chroot to host: %v", err),
			LogOutput:   output,
		}, err
	}

	// Step 4: Verify runtime.yml exists on host
	if streamer != nil {
		if err := streamer.SendProgress("✅ Verifying local installation..."); err != nil {
			log.Warn("failed to send progress", "error", err)
		}
	}

	hostRuntimeConfigPath := filepath.Join(hostRuntimePath, "runtime.yml")
	if _, err := ri.platform.Stat(hostRuntimeConfigPath); err != nil {
		return &RuntimeInstallResult{
			RuntimeSpec: runtimeSpec,
			Success:     false,
			Message:     "Installation completed but runtime.yml not found on host",
			LogOutput:   output,
		}, fmt.Errorf("runtime.yml not found on host at %s", hostRuntimeConfigPath)
	}

	if streamer != nil {
		if err := streamer.SendProgress("🎉 Local runtime installation completed successfully!"); err != nil {
			log.Warn("failed to send progress", "error", err)
		}
	}

	log.Info("local runtime installation completed successfully", "configPath", hostRuntimeConfigPath)
	return &RuntimeInstallResult{
		RuntimeSpec: runtimeSpec,
		Success:     true,
		Message:     "Runtime installed successfully from local files",
		InstallPath: hostRuntimeConfigPath,
		LogOutput:   output,
	}, nil
}

// writeFilesToChroot writes uploaded files to the chroot environment
func (ri *RuntimeInstaller) writeFilesToChroot(chrootDir, targetDir string, files []*RuntimeFile) error {
	fullTargetDir := filepath.Join(chrootDir, strings.TrimPrefix(targetDir, "/"))

	// Create target directory
	if err := os.MkdirAll(fullTargetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Write each file
	for _, file := range files {
		targetPath := filepath.Join(fullTargetDir, file.Path)

		// Create parent directory if needed
		parentDir := filepath.Dir(targetPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create parent directory for %s: %w", file.Path, err)
		}

		// Write file content
		var mode os.FileMode = 0644
		if file.Executable {
			mode = 0755
		}

		if err := os.WriteFile(targetPath, file.Content, mode); err != nil {
			return fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}

		ri.logger.Debug("wrote runtime file", "path", file.Path, "size", len(file.Content), "executable", file.Executable)
	}

	ri.logger.Info("wrote runtime files to chroot", "count", len(files), "targetDir", targetDir)
	return nil
}

// executeLocalSetupInChrootWithStreaming executes local setup script with streaming support
func (ri *RuntimeInstaller) executeLocalSetupInChrootWithStreaming(ctx context.Context, chrootDir, workDir, runtimeSpec string, streamer RuntimeInstallationStreamer) (string, error) {
	log := ri.logger.WithFields("workDir", workDir, "runtimeSpec", runtimeSpec)
	log.Info("executing local setup script in chroot with streaming")

	// Set environment variables for the setup script
	env := []string{
		ri.buildPathFromConfig(),
		fmt.Sprintf("RUNTIME_SPEC=%s", runtimeSpec),
		"RUNTIME_DIR=/opt/joblet/runtimes",
		fmt.Sprintf("BUILD_ID=local-install-%d", time.Now().Unix()),
		"JOBLET_CHROOT=true",        // Let scripts know they're in chroot
		"JOBLET_INSTALL_MODE=local", // Local file installation
	}

	// Create a script that changes to the correct directory and runs setup
	setupScript := fmt.Sprintf(`#!/bin/bash
set -e
cd %s
# Look for various setup script patterns
if [ -f setup.sh ]; then
    echo "Executing setup.sh in %s"
    chmod +x setup.sh
    ./setup.sh
elif [ -f install.sh ]; then
    echo "Executing install.sh in %s"
    chmod +x install.sh
    ./install.sh
else
    echo "No setup script found (setup.sh or install.sh)"
    ls -la
    exit 1
fi
`, workDir, workDir, workDir)

	// Write the setup script to chroot
	setupScriptPath := filepath.Join(chrootDir, "run_setup.sh")
	if err := os.WriteFile(setupScriptPath, []byte(setupScript), 0755); err != nil {
		return "", fmt.Errorf("failed to write setup script: %w", err)
	}

	// Execute the setup script with streaming support
	output, err := ri.executeChrootCommandWithStreaming(ctx, chrootDir, "/bin/bash", []string{"/run_setup.sh"}, env, streamer)
	if err != nil {
		ri.logger.Error("setup script failed", "error", err, "output", output)
		return output, fmt.Errorf("setup script execution failed: %w", err)
	}

	ri.logger.Info("setup script executed successfully")
	return output, nil
}

// createBuilderChroot creates a builder chroot environment with full host OS access
// This provides the complete build environment while preventing host contamination
func (ri *RuntimeInstaller) createBuilderChroot() (string, func(), error) {
	chrootDir := filepath.Join("/tmp", fmt.Sprintf("builder-chroot-%d", time.Now().UnixNano()))

	if err := os.MkdirAll(chrootDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create builder chroot directory: %w", err)
	}

	ri.logger.Info("creating builder chroot with full host OS access", "chrootDir", chrootDir)

	var mountsToCleanup []string

	// Mount complete host filesystem excluding /opt/* and special filesystems
	// Read all directories from host root
	entries, err := os.ReadDir("/")
	if err != nil {
		return "", nil, fmt.Errorf("failed to read host root directory: %w", err)
	}

	for _, entry := range entries {
		entryName := entry.Name()
		hostPath := filepath.Join("/", entryName)

		// Check what type of entry this is
		info, err := entry.Info()
		if err != nil {
			ri.logger.Debug("failed to get entry info", "entry", entryName, "error", err)
			continue
		}

		// Handle directories and symlinks
		if !info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
			continue // Skip files that are neither directories nor symlinks
		}

		// Skip directories that should not be mounted
		switch entryName {
		case "opt":
			// Skip /opt entirely to prevent recursion
			ri.logger.Debug("skipping /opt to prevent recursion")
			continue
		case "proc", "sys", "dev":
			// Skip special filesystems - will be handled separately
			ri.logger.Debug("skipping special filesystem", "dir", entryName)
			continue
		case "tmp":
			// Skip /tmp - will be handled separately with isolated mount
			ri.logger.Debug("skipping /tmp - will be handled separately")
			continue
		case "var", "etc":
			// CRITICAL: Copy these directories instead of bind mounting to prevent host contamination
			ri.logger.Info("copying directory to prevent host contamination", "dir", entryName)
			targetPath := filepath.Join(chrootDir, entryName)
			// Use cp -a to preserve permissions and create a full copy
			ri.logger.Info("starting copy operation", "source", hostPath, "target", targetPath)
			cpCmd := exec.Command("cp", "-a", hostPath, targetPath)
			output, err := cpCmd.CombinedOutput()
			if err != nil {
				ri.logger.Error("failed to copy directory", "source", hostPath, "target", targetPath, "error", err, "output", string(output))
				// This is critical - if we can't copy, we should fail the installation
				return "", nil, fmt.Errorf("failed to copy %s for isolation: %w", entryName, err)
			} else {
				ri.logger.Info("successfully copied directory for isolation", "source", hostPath, "target", targetPath)
			}
			continue
		}

		// Check if path exists and is accessible
		if _, err := os.Lstat(hostPath); err != nil {
			ri.logger.Debug("skipping host path - not accessible", "path", hostPath, "error", err)
			continue
		}

		targetPath := filepath.Join(chrootDir, entryName)

		// Handle symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			// Read the symlink target
			linkTarget, err := os.Readlink(hostPath)
			if err != nil {
				ri.logger.Warn("failed to read symlink", "source", hostPath, "error", err)
				continue
			}

			// Create the symlink in chroot
			if err := os.Symlink(linkTarget, targetPath); err != nil {
				ri.logger.Warn("failed to create symlink", "source", hostPath, "target", targetPath, "linkTarget", linkTarget, "error", err)
			} else {
				ri.logger.Debug("created symlink", "source", hostPath, "target", targetPath, "linkTarget", linkTarget)
			}
		} else {
			// Handle directories with READ-ONLY bind mount for most directories
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				ri.logger.Warn("failed to create target directory", "dir", targetPath, "error", err)
				continue
			}

			// Directories that can be read-only (binaries, libraries)
			readOnlyDirs := map[string]bool{
				"bin":   true,
				"sbin":  true,
				"lib":   true,
				"lib32": true,
				"lib64": true,
				"usr":   true,
			}

			if readOnlyDirs[entryName] {
				// Mount as read-only to prevent contamination
				if err := ri.bindMountReadOnly(hostPath, targetPath); err != nil {
					ri.logger.Warn("failed to mount host directory as read-only", "source", hostPath, "target", targetPath, "error", err)
				} else {
					mountsToCleanup = append(mountsToCleanup, targetPath)
					ri.logger.Debug("mounted host directory as read-only", "source", hostPath, "target", targetPath)
				}
			} else {
				// Regular bind mount for other directories
				if err := ri.bindMount(hostPath, targetPath); err != nil {
					ri.logger.Warn("failed to mount host directory", "source", hostPath, "target", targetPath, "error", err)
				} else {
					mountsToCleanup = append(mountsToCleanup, targetPath)
					ri.logger.Debug("mounted host directory", "source", hostPath, "target", targetPath)
				}
			}
		}
	}

	// Skip /opt entirely to prevent recursion as per design specification
	// Builder chroot = Host filesystem (/) minus /opt/* to prevent recursion
	// Create empty /opt directory but don't mount anything from host /opt
	optDir := filepath.Join(chrootDir, "opt")
	if err := os.MkdirAll(optDir, 0755); err != nil {
		ri.logger.Warn("failed to create empty /opt directory in chroot", "error", err)
	} else {
		ri.logger.Debug("created empty /opt directory in builder chroot (no host /opt mounts per design)")
	}

	// Create isolated /tmp for build process
	buildTmpDir := filepath.Join("/tmp", fmt.Sprintf("builder-tmp-%d", time.Now().UnixNano()))
	if err := os.MkdirAll(buildTmpDir, 0755); err != nil {
		ri.logger.Warn("failed to create isolated tmp directory", "error", err)
	} else {
		chrootTmpDir := filepath.Join(chrootDir, "tmp")
		if err := os.MkdirAll(chrootTmpDir, 0755); err == nil {
			if err := ri.bindMount(buildTmpDir, chrootTmpDir); err != nil {
				ri.logger.Warn("failed to mount isolated tmp", "error", err)
			} else {
				mountsToCleanup = append(mountsToCleanup, chrootTmpDir)
				ri.logger.Debug("mounted isolated tmp directory", "source", buildTmpDir, "target", chrootTmpDir)
			}
		}
	}

	// Create /opt/joblet/runtimes as writable directory inside chroot
	// Don't bind mount from host since the host directory may not exist yet
	jobletOptDir := filepath.Join(chrootDir, "opt/joblet")
	if err := os.MkdirAll(jobletOptDir, 0755); err != nil {
		ri.logger.Warn("failed to create /opt/joblet in chroot", "error", err)
	} else {
		// Create /opt/joblet/runtimes as a regular writable directory in chroot
		chrootRuntimeDir := filepath.Join(jobletOptDir, "runtimes")
		if err := os.MkdirAll(chrootRuntimeDir, 0755); err != nil {
			ri.logger.Warn("failed to create runtimes directory in chroot", "error", err)
		} else {
			ri.logger.Debug("created writable runtimes directory in chroot", "path", chrootRuntimeDir)
		}

		// Note: Runtime scripts are uploaded to /tmp in chroot (isolated tmp mount)
	}

	// Mount special filesystems for complete chroot environment
	specialMounts := []struct {
		source string
		target string
		fstype string
	}{
		{"proc", "proc", "proc"},
		{"sysfs", "sys", "sysfs"},
		{"/dev", "dev", ""},             // bind mount
		{"devpts", "dev/pts", "devpts"}, // pseudo-terminal support for APT/dpkg
	}

	for _, mount := range specialMounts {
		targetPath := filepath.Join(chrootDir, mount.target)
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			ri.logger.Warn("failed to create special mount directory", "target", mount.target, "error", err)
			continue
		}

		if mount.fstype != "" {
			// Mount proc and sysfs
			err := ri.mountSpecialFS(mount.source, targetPath, mount.fstype)
			if err != nil {
				ri.logger.Warn("failed to mount special filesystem", "type", mount.fstype, "target", targetPath, "error", err)
			} else {
				mountsToCleanup = append(mountsToCleanup, targetPath)
				ri.logger.Debug("mounted special filesystem", "type", mount.fstype, "target", targetPath)
			}
		} else {
			// Bind mount /dev
			if err := ri.bindMount(mount.source, targetPath); err != nil {
				ri.logger.Warn("failed to bind mount", "source", mount.source, "target", targetPath, "error", err)
			} else {
				mountsToCleanup = append(mountsToCleanup, targetPath)
				ri.logger.Debug("bind mounted", "source", mount.source, "target", targetPath)
			}
		}
	}

	// Create additional device nodes if needed
	if err := ri.createDeviceNodes(chrootDir); err != nil {
		ri.logger.Warn("failed to create additional device nodes", "error", err)
	}

	cleanup := func() {
		ri.logger.Debug("cleaning up builder chroot", "chrootDir", chrootDir)

		// Unmount in reverse order
		for i := len(mountsToCleanup) - 1; i >= 0; i-- {
			if err := ri.unmount(mountsToCleanup[i]); err != nil {
				ri.logger.Warn("failed to unmount", "path", mountsToCleanup[i], "error", err)
			}
		}

		// Clean up isolated tmp directory
		if buildTmpDir != "" {
			os.RemoveAll(buildTmpDir)
		}

		// Remove chroot directory
		os.RemoveAll(chrootDir)
		ri.logger.Debug("builder chroot cleanup completed", "chrootDir", chrootDir)
	}

	ri.logger.Info("builder chroot environment created with full host OS access", "chrootDir", chrootDir, "mountCount", len(mountsToCleanup))
	return chrootDir, cleanup, nil
}

// executeInstallationInChrootWithStreaming executes the runtime installation with optional streaming support
func (ri *RuntimeInstaller) executeInstallationInChrootWithStreaming(ctx context.Context, chrootDir, repository, branch, path, runtimeSpec string, streamer RuntimeInstallationStreamer) (*RuntimeInstallResult, error) {
	log := ri.logger.WithField("chrootDir", chrootDir)

	log.Info("executing runtime installation directly in chroot with streaming support")

	// Send initial progress if streamer available
	if streamer != nil {
		if err := streamer.SendProgress("🔧 Starting runtime installation..."); err != nil {
			log.Warn("failed to send progress", "error", err)
		}
	}

	// Step 1: Clone repository in chroot
	if streamer != nil {
		if err := streamer.SendProgress("📦 Cloning repository..."); err != nil {
			log.Warn("failed to send progress", "error", err)
		}
	}

	// Try git-free download first (more reliable, no git dependency)
	if err := ri.downloadRepositoryWithoutGit(ctx, chrootDir, repository, branch, streamer); err != nil {
		ri.logger.Info("git-free download failed, falling back to git clone", "error", err)
		if streamer != nil {
			_ = streamer.SendProgress("ZIP download failed, trying git clone as fallback...")
		}

		// Fallback to git clone
		if gitErr := ri.cloneRepositoryInChrootWithStreaming(ctx, chrootDir, repository, branch, streamer); gitErr != nil {
			return &RuntimeInstallResult{
				RuntimeSpec: runtimeSpec,
				Success:     false,
				Message:     fmt.Sprintf("Repository download failed: ZIP error: %v, Git error: %v", err, gitErr),
			}, gitErr
		}
	}

	// Step 2: Execute setup script in chroot
	if streamer != nil {
		if err := streamer.SendProgress("🏗️  Running setup script..."); err != nil {
			log.Warn("failed to send progress", "error", err)
		}
	}

	output, err := ri.executeSetupInChrootWithStreaming(ctx, chrootDir, path, runtimeSpec, streamer)
	if err != nil {
		return &RuntimeInstallResult{
			RuntimeSpec: runtimeSpec,
			Success:     false,
			Message:     fmt.Sprintf("Setup execution failed: %v", err),
			LogOutput:   output,
		}, err
	}

	// Step 3: Verify installation
	if streamer != nil {
		if err := streamer.SendProgress("✅ Verifying installation..."); err != nil {
			log.Warn("failed to send progress", "error", err)
		}
	}

	runtimeConfigPath := filepath.Join(ri.config.Runtime.BasePath, ri.getRuntimePath(runtimeSpec), "runtime.yml")
	if _, err := ri.platform.Stat(runtimeConfigPath); err != nil {
		return &RuntimeInstallResult{
			RuntimeSpec: runtimeSpec,
			Success:     false,
			Message:     "Installation completed but runtime.yml not found",
			LogOutput:   output,
		}, fmt.Errorf("runtime.yml not created at %s", runtimeConfigPath)
	}

	if streamer != nil {
		if err := streamer.SendProgress("🎉 Runtime installation completed successfully!"); err != nil {
			log.Warn("failed to send progress", "error", err)
		}
	}

	log.Info("runtime installation completed successfully", "configPath", runtimeConfigPath)

	return &RuntimeInstallResult{
		RuntimeSpec: runtimeSpec,
		Success:     true,
		Message:     "Runtime installed successfully",
		InstallPath: runtimeConfigPath,
		LogOutput:   output,
	}, nil
}

// cloneRepositoryInChrootWithStreaming clones the git repository with streaming support
func (ri *RuntimeInstaller) cloneRepositoryInChrootWithStreaming(ctx context.Context, chrootDir, repository, branch string, streamer RuntimeInstallationStreamer) error {
	log := ri.logger.WithFields("repository", repository, "branch", branch)
	log.Info("cloning repository in chroot with streaming")

	// Build git clone command arguments
	var gitArgs []string
	if branch != "" {
		gitArgs = []string{"clone", "--depth", "1", "--branch", branch, fmt.Sprintf("https://github.com/%s.git", repository), "/tmp/runtime-build"}
	} else {
		gitArgs = []string{"clone", "--depth", "1", fmt.Sprintf("https://github.com/%s.git", repository), "/tmp/runtime-build"}
	}

	env := []string{
		ri.buildPathFromConfig(),
		"HOME=/root",
	}

	// Execute git clone with streaming support
	output, err := ri.executeChrootCommandWithStreaming(ctx, chrootDir, "git", gitArgs, env, streamer)
	if err != nil {
		log.Error("git clone failed", "error", err, "output", output)
		return fmt.Errorf("git clone failed: %w", err)
	}

	log.Info("repository cloned successfully")
	return nil
}

// downloadRepositoryWithoutGit downloads repository as ZIP (git-free alternative)
func (ri *RuntimeInstaller) downloadRepositoryWithoutGit(ctx context.Context, chrootDir, repository, branch string, streamer RuntimeInstallationStreamer) error {
	log := ri.logger.WithFields("repository", repository, "branch", branch)
	log.Info("downloading repository without git (using ZIP download)")

	if streamer != nil {
		_ = streamer.SendProgress(fmt.Sprintf("Downloading repository %s (branch: %s) without git...", repository, branch))
	}

	// GitHub ZIP download URL format: https://github.com/owner/repo/archive/refs/heads/branch.zip
	zipURL := fmt.Sprintf("https://github.com/%s/archive/refs/heads/%s.zip", repository, branch)

	// Download to chroot environment
	zipPath := filepath.Join(chrootDir, "tmp", "runtime-repo.zip")
	extractDir := filepath.Join(chrootDir, "tmp", "runtime-build")

	// Create tmp directory in chroot
	tmpDir := filepath.Join(chrootDir, "tmp")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("failed to create tmp directory: %w", err)
	}

	// Download the ZIP file
	resp, err := http.Get(zipURL)
	if err != nil {
		return fmt.Errorf("failed to download repository ZIP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download repository ZIP: HTTP %d", resp.StatusCode)
	}

	// Create the ZIP file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create ZIP file: %w", err)
	}
	defer zipFile.Close()

	// Copy the downloaded content
	_, err = io.Copy(zipFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write ZIP file: %w", err)
	}

	if streamer != nil {
		_ = streamer.SendProgress("Repository ZIP downloaded successfully")
		_ = streamer.SendProgress("Extracting repository...")
	}

	// Extract the ZIP file
	if err := ri.extractZipFile(zipPath, extractDir, repository, branch); err != nil {
		return fmt.Errorf("failed to extract ZIP: %w", err)
	}

	// Clean up the ZIP file
	os.Remove(zipPath)

	if streamer != nil {
		_ = streamer.SendProgress("Repository extracted successfully")
	}

	log.Info("repository downloaded and extracted successfully")
	return nil
}

// extractZipFile extracts a ZIP file to the target directory
func (ri *RuntimeInstaller) extractZipFile(zipPath, extractDir, repository, branch string) error {
	// Open the ZIP file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Create extract directory
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return err
	}

	// Expected folder name in ZIP (GitHub uses repo-branch format)
	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format: %s", repository)
	}
	repoName := parts[1]
	expectedPrefix := fmt.Sprintf("%s-%s/", repoName, branch)

	// Extract files
	for _, file := range reader.File {
		// Skip directories and files not in the expected prefix
		if file.FileInfo().IsDir() || !strings.HasPrefix(file.Name, expectedPrefix) {
			continue
		}

		// Remove the prefix to get the relative path
		relativePath := strings.TrimPrefix(file.Name, expectedPrefix)
		if relativePath == "" {
			continue
		}

		targetPath := filepath.Join(extractDir, relativePath)

		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(targetPath), file.FileInfo().Mode()); err != nil {
			return err
		}

		// Open file in ZIP
		zipFileReader, err := file.Open()
		if err != nil {
			return err
		}

		// Create target file
		targetFile, err := os.Create(targetPath)
		if err != nil {
			zipFileReader.Close()
			return err
		}

		// Copy content
		_, err = io.Copy(targetFile, zipFileReader)
		zipFileReader.Close()
		targetFile.Close()

		if err != nil {
			return err
		}

		// Set file permissions
		if err := os.Chmod(targetPath, file.FileInfo().Mode()); err != nil {
			return err
		}
	}

	return nil
}

// executeSetupInChrootWithStreaming executes the setup script with streaming support
func (ri *RuntimeInstaller) executeSetupInChrootWithStreaming(ctx context.Context, chrootDir, path, runtimeSpec string, streamer RuntimeInstallationStreamer) (string, error) {
	log := ri.logger.WithFields("path", path, "runtimeSpec", runtimeSpec)
	log.Info("executing setup script in chroot with streaming")

	// Determine working directory
	workDir := "/tmp/runtime-build"
	if path != "" {
		workDir = fmt.Sprintf("/tmp/runtime-build/%s", path)
	}

	// Set environment variables for the setup script
	env := []string{
		ri.buildPathFromConfig(),
		fmt.Sprintf("RUNTIME_SPEC=%s", runtimeSpec),
		"RUNTIME_DIR=/opt/joblet/runtimes",
		fmt.Sprintf("BUILD_ID=install-%d", time.Now().Unix()),
		"JOBLET_CHROOT=true",         // Let scripts know they're in chroot
		"JOBLET_INSTALL_MODE=direct", // Direct installation (not via job system)
	}

	// Create a script that changes to the correct directory and runs setup
	setupScript := fmt.Sprintf(`#!/bin/bash
set -e
cd %s
# Look for various setup script patterns
if [ -f setup.sh ]; then
    echo "Executing setup.sh in %s"
    chmod +x setup.sh
    ./setup.sh
elif [ -f install.sh ]; then
    echo "Executing install.sh in %s"
    chmod +x install.sh
    ./install.sh
else
    echo "No setup script found in %s"
    exit 1
fi
echo "Setup completed successfully"`, workDir, workDir, workDir, workDir)

	// Write the setup script to chroot
	setupScriptPath := filepath.Join(chrootDir, "tmp", "run-setup.sh")
	if err := ri.platform.WriteFile(setupScriptPath, []byte(setupScript), 0755); err != nil {
		return "", fmt.Errorf("failed to write setup script: %w", err)
	}

	// Execute setup script with streaming
	output, err := ri.executeChrootCommandWithStreaming(ctx, chrootDir, "/bin/bash", []string{"/tmp/run-setup.sh"}, env, streamer)
	if err != nil {
		log.Error("setup script execution failed", "error", err, "output", output)
		return output, fmt.Errorf("setup script execution failed: %w", err)
	}

	return output, nil
}

// createDeviceNodes creates essential device nodes in the chroot environment
func (ri *RuntimeInstaller) createDeviceNodes(chrootDir string) error {
	devDir := filepath.Join(chrootDir, "dev")

	// Device nodes required by git and other tools
	devices := []struct {
		name  string
		major uint32
		minor uint32
		mode  uint32
	}{
		{"null", 1, 3, syscall.S_IFCHR | 0666},
		{"zero", 1, 5, syscall.S_IFCHR | 0666},
		{"random", 1, 8, syscall.S_IFCHR | 0644},
		{"urandom", 1, 9, syscall.S_IFCHR | 0644},
	}

	for _, device := range devices {
		devicePath := filepath.Join(devDir, device.name)
		dev := int(ri.makedev(device.major, device.minor))

		if err := syscall.Mknod(devicePath, device.mode, dev); err != nil {
			ri.logger.Warn("failed to create device node", "device", device.name, "path", devicePath, "error", err)
			continue
		}

	}

	return nil
}

// makedev creates a device number from major and minor numbers
func (ri *RuntimeInstaller) makedev(major, minor uint32) uint64 {
	return uint64(major<<8 | minor)
}

// bindMount performs a bind mount
func (ri *RuntimeInstaller) bindMount(source, target string) error {
	flags := uintptr(syscall.MS_BIND)
	return syscall.Mount(source, target, "", flags, "")
}

// bindMountReadOnly performs a read-only bind mount
func (ri *RuntimeInstaller) bindMountReadOnly(source, target string) error {
	// First do the bind mount
	flags := uintptr(syscall.MS_BIND)
	if err := syscall.Mount(source, target, "", flags, ""); err != nil {
		return err
	}
	// Then remount as read-only
	flags = uintptr(syscall.MS_BIND | syscall.MS_REMOUNT | syscall.MS_RDONLY)
	return syscall.Mount(source, target, "", flags, "")
}

// mountSpecialFS mounts special filesystems like proc, sysfs
func (ri *RuntimeInstaller) mountSpecialFS(source, target, fstype string) error {
	return syscall.Mount(source, target, fstype, 0, "")
}

// unmount unmounts a filesystem
func (ri *RuntimeInstaller) unmount(target string) error {
	return syscall.Unmount(target, 0)
}

// autoDetectRuntimePath determines the runtime path from the spec
func (ri *RuntimeInstaller) autoDetectRuntimePath(runtimeSpec string) string {
	// Parse runtime spec like "python:3.11-ml" -> "python-3.11-ml"
	parts := strings.SplitN(runtimeSpec, ":", 2)
	if len(parts) != 2 {
		return runtimeSpec // fallback
	}

	language := parts[0]
	version := parts[1]

	// Use flat structure as per design
	return fmt.Sprintf("%s-%s", language, version)
}

// getRuntimePath gets the expected runtime installation path
func (ri *RuntimeInstaller) getRuntimePath(runtimeSpec string) string {
	// Runtime name directly maps to directory name (e.g., "openjdk-21" -> "/opt/joblet/runtimes/openjdk-21")
	return runtimeSpec
}

// executeChrootCommandWithStreaming executes a command in chroot environment with optional streaming support
func (ri *RuntimeInstaller) executeChrootCommandWithStreaming(ctx context.Context, chrootDir, command string, args []string, env []string, streamer RuntimeInstallationStreamer) (string, error) {
	log := ri.logger.WithFields("command", command, "args", args)
	log.Debug("executing command in chroot using proper chroot syscall")

	// Use the system's chroot command to properly chroot and execute
	// This ensures the command sees the chroot as the root filesystem
	chrootArgs := []string{"chroot", chrootDir, command}
	chrootArgs = append(chrootArgs, args...)

	// Create the command
	cmd := exec.CommandContext(ctx, chrootArgs[0], chrootArgs[1:]...)
	cmd.Env = env

	var outputStr string
	var err error

	if streamer != nil {
		// Stream output in real-time
		outputStr, err = ri.runCommandWithStreaming(cmd, streamer)
	} else {
		// Use legacy combined output method
		output, cmdErr := cmd.CombinedOutput()
		outputStr = string(output)
		err = cmdErr
	}

	if err != nil {
		log.Debug("chroot command failed", "error", err, "output", outputStr)
		return outputStr, fmt.Errorf("command execution failed: %w", err)
	}

	log.Debug("chroot command executed successfully")
	return outputStr, nil
}

// runCommandWithStreaming executes a command and streams output in real-time
func (ri *RuntimeInstaller) runCommandWithStreaming(cmd *exec.Cmd, streamer RuntimeInstallationStreamer) (string, error) {
	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Create a channel to collect output
	outputChan := make(chan string, 100)
	done := make(chan error, 2)

	// Stream stdout
	go func() {
		defer func() { done <- nil }()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			outputChan <- line + "\n"

			// Send to streamer if available
			if streamer != nil {
				if err := streamer.SendLog([]byte(line + "\n")); err != nil {
					ri.logger.Warn("failed to send log to streamer", "error", err)
				}
			}
		}
	}()

	// Stream stderr
	go func() {
		defer func() { done <- nil }()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			outputChan <- line + "\n"

			// Send to streamer if available
			if streamer != nil {
				if err := streamer.SendLog([]byte(line + "\n")); err != nil {
					ri.logger.Warn("failed to send log to streamer", "error", err)
				}
			}
		}
	}()

	// Collect all output
	var allOutput strings.Builder
	outputDone := make(chan struct{})

	go func() {
		for output := range outputChan {
			allOutput.WriteString(output)
		}
		close(outputDone)
	}()

	// Wait for streaming goroutines to complete
	<-done
	<-done

	// Close output channel and wait for collection to finish
	close(outputChan)
	<-outputDone

	// Wait for command to finish
	err = cmd.Wait()
	return allOutput.String(), err
}

// findLocalRuntime searches for a local runtime in the development environment
func (ri *RuntimeInstaller) findLocalRuntime(runtimeSpec string) string {
	// Parse runtime spec
	parts := strings.SplitN(runtimeSpec, ":", 2)
	if len(parts) != 2 {
		return ""
	}

	// Build possible paths - check environment variables and common locations
	possiblePaths := []string{}

	// Check JOBLET_DEV_PATH environment variable if set
	if devPath := os.Getenv("JOBLET_DEV_PATH"); devPath != "" {
		possiblePaths = append(possiblePaths,
			fmt.Sprintf("%s/runtimes/%s-%s", devPath, parts[0], parts[1]))
	}

	// Add common development paths
	possiblePaths = append(possiblePaths,
		fmt.Sprintf("/home/jay/joblet/runtimes/%s-%s", parts[0], parts[1]), // Default dev path
		fmt.Sprintf("./runtimes/%s-%s", parts[0], parts[1]),                // Current directory
		fmt.Sprintf("../runtimes/%s-%s", parts[0], parts[1]),               // Parent directory
		fmt.Sprintf("/opt/joblet/dev/runtimes/%s-%s", parts[0], parts[1]),  // Alternative dev path
	)

	for _, path := range possiblePaths {
		if ri.platform.DirExists(path) {
			// Look for various setup script patterns
			setupScripts := []string{"setup.sh", "install.sh"}

			// Look for runtime-specific scripts too
			if entries, err := ri.platform.ReadDir(path); err == nil {
				for _, entry := range entries {
					name := entry.Name()
					if strings.HasPrefix(name, "setup_") && strings.HasSuffix(name, ".sh") {
						setupScripts = append(setupScripts, name)
					}
					if strings.HasPrefix(name, "install_") && strings.HasSuffix(name, ".sh") {
						setupScripts = append(setupScripts, name)
					}
				}
			}

			// Check if any setup script exists
			for _, scriptName := range setupScripts {
				setupScript := filepath.Join(path, scriptName)
				if _, err := ri.platform.Stat(setupScript); err == nil {
					return path
				}
			}
		}
	}

	return ""
}
