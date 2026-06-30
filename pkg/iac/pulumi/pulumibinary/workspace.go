package pulumibinary

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/fileutil"
)

const (
	// PulumiYamlTemplate is the template for generating Pulumi.yaml with binary option
	PulumiYamlTemplate = `name: %s
runtime:
  name: go
  options:
    binary: %s
description: Auto-generated workspace for %s binary execution
`
)

// WorkspaceResult contains the result of setting up a binary workspace
type WorkspaceResult struct {
	// WorkspacePath is the path to the workspace directory containing Pulumi.yaml
	WorkspacePath string
	// BinaryPath is the path to the binary being used
	BinaryPath string
	// CleanupFunc is a function to clean up the workspace
	CleanupFunc func() error
	// ShouldCleanup indicates if cleanup should be performed
	ShouldCleanup bool
}

// GetWorkspaceDir returns the base directory for binary workspaces
// (~/.planton/pulumi/workspaces/)
func GetWorkspaceDir() (string, error) {
	pulumiBaseDir, err := GetPulumiBaseDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get Pulumi base directory")
	}
	return filepath.Join(pulumiBaseDir, WorkspacesSubDir), nil
}

// GetStackWorkspaceDir returns the workspace directory for a specific stack
// (~/.planton/pulumi/workspaces/{stack-fqdn}/)
func GetStackWorkspaceDir(stackFqdn string) (string, error) {
	baseDir, err := GetWorkspaceDir()
	if err != nil {
		return "", err
	}

	// Sanitize stack FQDN for directory name
	// Replace slashes with dashes for filesystem compatibility
	sanitized := strings.ReplaceAll(stackFqdn, "/", "-")
	return filepath.Join(baseDir, sanitized), nil
}

// SetupBinaryWorkspace creates a minimal workspace directory for binary execution.
// It generates a Pulumi.yaml that references the pre-built binary.
// Returns the workspace path that can be used with pulumi CLI.
func SetupBinaryWorkspace(binaryPath, stackFqdn, componentName string) (*WorkspaceResult, error) {
	// Get workspace directory for this stack
	workspacePath, err := GetStackWorkspaceDir(stackFqdn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get stack workspace directory")
	}

	// Create workspace directory if it doesn't exist
	if !fileutil.IsDirExists(workspacePath) {
		if err := os.MkdirAll(workspacePath, 0755); err != nil {
			return nil, errors.Wrapf(err, "failed to create workspace directory %s", workspacePath)
		}
	}

	// Generate project name from component
	projectName := strings.ToLower(componentName)

	// Generate Pulumi.yaml content
	pulumiYamlContent := GeneratePulumiYaml(binaryPath, projectName, componentName)

	// Write Pulumi.yaml
	pulumiYamlPath := filepath.Join(workspacePath, "Pulumi.yaml")
	if err := os.WriteFile(pulumiYamlPath, []byte(pulumiYamlContent), 0644); err != nil {
		return nil, errors.Wrapf(err, "failed to write Pulumi.yaml to %s", pulumiYamlPath)
	}

	// Create cleanup function
	cleanupFunc := func() error {
		return CleanupWorkspace(workspacePath)
	}

	return &WorkspaceResult{
		WorkspacePath: workspacePath,
		BinaryPath:    binaryPath,
		CleanupFunc:   cleanupFunc,
		ShouldCleanup: true,
	}, nil
}

// GeneratePulumiYaml generates the content for Pulumi.yaml with binary option
func GeneratePulumiYaml(binaryPath, projectName, componentName string) string {
	return fmt.Sprintf(PulumiYamlTemplate, projectName, binaryPath, componentName)
}

// CleanupWorkspace removes a workspace directory
func CleanupWorkspace(workspacePath string) error {
	if workspacePath == "" {
		return nil
	}

	// Safety check: ensure we're only deleting within our workspace directory
	baseDir, err := GetWorkspaceDir()
	if err != nil {
		return errors.Wrap(err, "failed to get base workspace directory for safety check")
	}

	if !strings.HasPrefix(workspacePath, baseDir) {
		return errors.Errorf("refusing to delete directory outside of workspace: %s", workspacePath)
	}

	if !fileutil.IsDirExists(workspacePath) {
		return nil // Already cleaned up
	}

	if err := os.RemoveAll(workspacePath); err != nil {
		return errors.Wrapf(err, "failed to cleanup workspace at %s", workspacePath)
	}

	return nil
}

// UpdateProjectName updates the project name in an existing Pulumi.yaml
// This is needed because Pulumi stack names are tied to project names
func UpdateProjectName(workspacePath, newProjectName string) error {
	pulumiYamlPath := filepath.Join(workspacePath, "Pulumi.yaml")

	content, err := os.ReadFile(pulumiYamlPath)
	if err != nil {
		return errors.Wrapf(err, "failed to read Pulumi.yaml from %s", pulumiYamlPath)
	}

	// Simple replacement - look for name: line and replace it
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "name:") {
			lines[i] = fmt.Sprintf("name: %s", newProjectName)
			break
		}
	}

	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(pulumiYamlPath, []byte(newContent), 0644); err != nil {
		return errors.Wrapf(err, "failed to write updated Pulumi.yaml to %s", pulumiYamlPath)
	}

	return nil
}
