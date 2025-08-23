package installers

import (
	"context"
	"fmt"
	pb "joblet/api/gen"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GitHubInstaller handles runtime installation from GitHub repositories
type GitHubInstaller struct {
	*BaseInstaller
}

// NewGitHubInstaller creates a new GitHub runtime installer
func NewGitHubInstaller() *GitHubInstaller {
	return &GitHubInstaller{
		BaseInstaller: NewBaseInstaller("github_install.sh"),
	}
}

// Install installs a runtime from a GitHub repository
func (g *GitHubInstaller) Install(ctx context.Context, spec *InstallSpec) (*InstallResult, error) {
	github, ok := spec.Source.SourceType.(*pb.RuntimeSourceConfig_Github)
	if !ok {
		return nil, fmt.Errorf("invalid source type for GitHub installer")
	}

	// Set default values
	repository := github.Github.Repository
	if repository == "" {
		repository = "joblet-org/runtimes"
	}

	branch := github.Github.Branch
	if branch == "" {
		branch = "main"
	}

	path := github.Github.Path
	if path == "" {
		path = ""
	}

	// Validate repository format
	repositoryParts := strings.Split(repository, "/")
	if len(repositoryParts) != 2 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid repository format: %s", repository)
	}

	// Prepare template data
	templateData := &TemplateData{
		RuntimeSpec: spec.RuntimeSpec,
		Repository:  repository,
		Branch:      branch,
		Path:        path,
		BuildArgs:   spec.BuildArgs,
	}

	// Render the installation script
	script, err := g.RenderTemplate(templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate GitHub install script: %w", err)
	}

	return g.CreateInstallResult(script, true, fmt.Sprintf("GitHub installation script generated for %s", spec.RuntimeSpec)), nil
}

// Validate checks if the GitHub source configuration is valid
func (g *GitHubInstaller) Validate(source *pb.RuntimeSourceConfig) error {
	github, ok := source.SourceType.(*pb.RuntimeSourceConfig_Github)
	if !ok {
		return fmt.Errorf("invalid source type for GitHub installer")
	}

	// Repository format validation
	if github.Github.Repository != "" {
		repositoryParts := strings.Split(github.Github.Repository, "/")
		if len(repositoryParts) != 2 {
			return fmt.Errorf("invalid repository format: %s (expected: owner/repo)", github.Github.Repository)
		}
	}

	return nil
}

// GetSourceType returns the source type handled by this installer
func (g *GitHubInstaller) GetSourceType() string {
	return "github"
}
