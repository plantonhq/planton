package staging

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/workspace"
	"github.com/plantonhq/planton/pkg/fileutil"
	"github.com/plantonhq/planton/pkg/iac/gitrepo"
)

const (
	stagingDirName  = "staging"
	versionFileName = ".version"
)

// GetStagingDir returns the path to the staging directory (~/.planton/staging)
func GetStagingDir() (string, error) {
	workspaceDir, err := workspace.GetWorkspaceDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get workspace directory")
	}
	return filepath.Join(workspaceDir, stagingDirName), nil
}

// GetStagingRepoPath returns the path to the cloned repository in staging
// (~/.planton/staging/planton)
func GetStagingRepoPath() (string, error) {
	stagingDir, err := GetStagingDir()
	if err != nil {
		return "", err
	}

	repoName, err := gitrepo.ExtractRepoName(gitrepo.CloneUrl)
	if err != nil {
		return "", errors.Wrapf(err, "failed to extract repo name from %s", gitrepo.CloneUrl)
	}

	return filepath.Join(stagingDir, repoName), nil
}

// GetVersionFilePath returns the path to the version tracking file
func GetVersionFilePath() (string, error) {
	stagingDir, err := GetStagingDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(stagingDir, versionFileName), nil
}

// GetCurrentStagingVersion reads the currently checked out version from the version file
func GetCurrentStagingVersion() (string, error) {
	versionFile, err := GetVersionFilePath()
	if err != nil {
		return "", err
	}

	if !fileutil.IsDirExists(filepath.Dir(versionFile)) {
		return "", nil // Staging doesn't exist yet
	}

	content, err := os.ReadFile(versionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", errors.Wrap(err, "failed to read version file")
	}

	return strings.TrimSpace(string(content)), nil
}

// WriteCurrentStagingVersion writes the version to the version file
func WriteCurrentStagingVersion(version string) error {
	versionFile, err := GetVersionFilePath()
	if err != nil {
		return err
	}

	// Ensure staging directory exists
	stagingDir := filepath.Dir(versionFile)
	if !fileutil.IsDirExists(stagingDir) {
		if err := os.MkdirAll(stagingDir, os.ModePerm); err != nil {
			return errors.Wrapf(err, "failed to create staging directory %s", stagingDir)
		}
	}

	if err := os.WriteFile(versionFile, []byte(version+"\n"), 0644); err != nil {
		return errors.Wrap(err, "failed to write version file")
	}

	return nil
}

// StagingExists checks if the staging repository already exists
func StagingExists() (bool, error) {
	repoPath, err := GetStagingRepoPath()
	if err != nil {
		return false, err
	}

	// Check if it's a valid git repository
	gitDir := filepath.Join(repoPath, ".git")
	return fileutil.IsDirExists(gitDir), nil
}

// EnsureStaging ensures the staging repository exists and is at the correct version.
// If staging doesn't exist, it clones the repository.
// If the version doesn't match, it checks out the correct version.
func EnsureStaging(targetVersion string) error {
	exists, err := StagingExists()
	if err != nil {
		return errors.Wrap(err, "failed to check if staging exists")
	}

	repoPath, err := GetStagingRepoPath()
	if err != nil {
		return err
	}

	if !exists {
		// Clone the repository
		if err := cloneToStaging(); err != nil {
			return errors.Wrap(err, "failed to clone repository to staging")
		}
	}

	// Check current version
	currentVersion, err := GetCurrentStagingVersion()
	if err != nil {
		return errors.Wrap(err, "failed to get current staging version")
	}

	// If version matches, we're done
	if currentVersion == targetVersion && targetVersion != "" {
		return nil
	}

	// Checkout the target version if specified
	if targetVersion != "" {
		if err := checkoutVersion(repoPath, targetVersion); err != nil {
			return errors.Wrapf(err, "failed to checkout version %s", targetVersion)
		}

		if err := WriteCurrentStagingVersion(targetVersion); err != nil {
			return errors.Wrap(err, "failed to write version file")
		}
	}

	return nil
}

// cloneToStaging clones the repository to the staging directory
func cloneToStaging() error {
	stagingDir, err := GetStagingDir()
	if err != nil {
		return err
	}

	// Ensure staging directory exists
	if !fileutil.IsDirExists(stagingDir) {
		if err := os.MkdirAll(stagingDir, os.ModePerm); err != nil {
			return errors.Wrapf(err, "failed to create staging directory %s", stagingDir)
		}
	}

	repoPath, err := GetStagingRepoPath()
	if err != nil {
		return err
	}

	cliprint.PrintStep("Cloning Planton repository to staging area...")
	cliprint.PrintInfo("This is a one-time operation. Future executions will use local copy.")

	cmd := exec.Command("git", "clone", "--progress", gitrepo.CloneUrl, repoPath)
	var cloneStderr bytes.Buffer
	cmd.Stderr = &cloneStderr

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to clone repository: %s", cloneStderr.String())
	}

	cliprint.PrintSuccess(fmt.Sprintf("Repository cloned to: %s", repoPath))
	return nil
}

// checkoutVersion checks out a specific version/tag/branch in the repository
func checkoutVersion(repoPath, version string) error {
	// First fetch to ensure we have all tags and branches
	fetchCmd := exec.Command("git", "-C", repoPath, "fetch", "--all", "--tags")
	var fetchStderr bytes.Buffer
	fetchCmd.Stderr = &fetchStderr
	if err := fetchCmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to fetch tags: %s", fetchStderr.String())
	}

	// Checkout the version
	checkoutCmd := exec.Command("git", "-C", repoPath, "checkout", version)
	var checkoutStderr bytes.Buffer
	checkoutCmd.Stderr = &checkoutStderr

	if err := checkoutCmd.Run(); err != nil {
		errMsg := checkoutStderr.String()
		if strings.Contains(errMsg, "did not match any") {
			return errors.Errorf("version '%s' not found. Ensure the tag, branch, or commit SHA exists", version)
		}
		return errors.Wrapf(err, "failed to checkout %s: %s", version, errMsg)
	}

	return nil
}

// Pull fetches and pulls the latest changes from upstream
func Pull() error {
	exists, err := StagingExists()
	if err != nil {
		return errors.Wrap(err, "failed to check if staging exists")
	}

	if !exists {
		// Clone first if staging doesn't exist
		if err := cloneToStaging(); err != nil {
			return err
		}
		return nil
	}

	repoPath, err := GetStagingRepoPath()
	if err != nil {
		return err
	}

	// Remember the current version to restore after pull
	previousVersion, _ := GetCurrentStagingVersion()

	cliprint.PrintStep("Fetching latest changes from upstream...")

	// Fetch all remotes
	fetchCmd := exec.Command("git", "-C", repoPath, "fetch", "--all", "--tags")
	var fetchStderr bytes.Buffer
	fetchCmd.Stderr = &fetchStderr
	if err := fetchCmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to fetch from upstream: %s", fetchStderr.String())
	}
	cliprint.PrintSuccess("Fetched all tags and branches")

	// Checkout main branch before pulling (handles detached HEAD state)
	cliprint.PrintStep("Checking out main branch...")
	checkoutCmd := exec.Command("git", "-C", repoPath, "checkout", "main")
	var checkoutStderr bytes.Buffer
	checkoutCmd.Stderr = &checkoutStderr
	if err := checkoutCmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to checkout main branch: %s", checkoutStderr.String())
	}

	// Pull changes on main
	cliprint.PrintStep("Pulling latest changes...")
	pullCmd := exec.Command("git", "-C", repoPath, "pull")
	var pullStderr bytes.Buffer
	pullCmd.Stderr = &pullStderr
	if err := pullCmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to pull from upstream: %s", pullStderr.String())
	}
	cliprint.PrintSuccess("Pulled latest changes from main")

	// If there was a previous version (not main), restore it
	if previousVersion != "" && previousVersion != "main" {
		cliprint.PrintStep(fmt.Sprintf("Restoring previous version: %s", previousVersion))
		if err := checkoutVersion(repoPath, previousVersion); err != nil {
			// If we can't restore, stay on main and update version file
			cliprint.PrintInfo(fmt.Sprintf("Could not restore version '%s', staying on main", previousVersion))
			if err := WriteCurrentStagingVersion("main"); err != nil {
				return errors.Wrap(err, "failed to update version file")
			}
			return nil
		}
		cliprint.PrintSuccess(fmt.Sprintf("Restored to version: %s", previousVersion))
		if err := WriteCurrentStagingVersion(previousVersion); err != nil {
			return errors.Wrap(err, "failed to update version file")
		}
	} else {
		// Update version file with main
		if err := WriteCurrentStagingVersion("main"); err != nil {
			return errors.Wrap(err, "failed to update version file")
		}
	}

	return nil
}

// getLatestTag fetches the latest tag from the repository
func getLatestTag(repoPath string) (string, error) {
	// First fetch all tags to ensure we have the latest
	fetchCmd := exec.Command("git", "-C", repoPath, "fetch", "--tags")
	var fetchStderr bytes.Buffer
	fetchCmd.Stderr = &fetchStderr
	if err := fetchCmd.Run(); err != nil {
		return "", errors.Wrapf(err, "failed to fetch tags: %s", fetchStderr.String())
	}

	// Get the latest tag sorted by version (semver-aware)
	// Using -v:refname sorts tags by version, and we take the last one
	tagCmd := exec.Command("git", "-C", repoPath, "tag", "--sort=-v:refname")
	tagOutput, err := tagCmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "failed to list tags")
	}

	tags := strings.Split(strings.TrimSpace(string(tagOutput)), "\n")
	if len(tags) == 0 || tags[0] == "" {
		return "", errors.New("no tags found in repository")
	}

	return tags[0], nil
}

// Checkout checks out a specific version in the staging area
func Checkout(version string) error {
	exists, err := StagingExists()
	if err != nil {
		return errors.Wrap(err, "failed to check if staging exists")
	}

	if !exists {
		// Clone first if staging doesn't exist
		if err := cloneToStaging(); err != nil {
			return err
		}
	}

	repoPath, err := GetStagingRepoPath()
	if err != nil {
		return err
	}

	// Handle "latest" as a special case - resolve to the latest tag
	if version == "latest" {
		cliprint.PrintStep("Resolving latest tag from repository...")
		latestTag, err := getLatestTag(repoPath)
		if err != nil {
			return errors.Wrap(err, "failed to resolve latest tag")
		}
		cliprint.PrintSuccess(fmt.Sprintf("Latest tag: %s", latestTag))
		version = latestTag
	}

	cliprint.PrintStep(fmt.Sprintf("Checking out version: %s", version))

	if err := checkoutVersion(repoPath, version); err != nil {
		return err
	}

	if err := WriteCurrentStagingVersion(version); err != nil {
		return errors.Wrap(err, "failed to update version file")
	}

	cliprint.PrintSuccess(fmt.Sprintf("Checked out version: %s", version))
	return nil
}

// getCurrentHead returns the current HEAD reference (tag if on tag, otherwise commit SHA)
func getCurrentHead(repoPath string) (string, error) {
	// Try to get current tag
	tagCmd := exec.Command("git", "-C", repoPath, "describe", "--tags", "--exact-match")
	tagOutput, err := tagCmd.Output()
	if err == nil {
		return strings.TrimSpace(string(tagOutput)), nil
	}

	// Fall back to commit SHA
	shaCmd := exec.Command("git", "-C", repoPath, "rev-parse", "--short", "HEAD")
	shaOutput, err := shaCmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "failed to get HEAD SHA")
	}

	return strings.TrimSpace(string(shaOutput)), nil
}

// CopyToWorkspace copies the staging repository to the specified destination directory.
// Returns the path to the copied repository.
func CopyToWorkspace(destDir string) (string, error) {
	repoPath, err := GetStagingRepoPath()
	if err != nil {
		return "", err
	}

	repoName, err := gitrepo.ExtractRepoName(gitrepo.CloneUrl)
	if err != nil {
		return "", errors.Wrapf(err, "failed to extract repo name")
	}

	destPath := filepath.Join(destDir, repoName)

	// Remove destination if it exists
	if fileutil.IsDirExists(destPath) {
		if err := os.RemoveAll(destPath); err != nil {
			return "", errors.Wrapf(err, "failed to remove existing directory %s", destPath)
		}
	}

	// Ensure destination parent directory exists
	if !fileutil.IsDirExists(destDir) {
		if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
			return "", errors.Wrapf(err, "failed to create destination directory %s", destDir)
		}
	}

	// Copy using cp -r for efficiency
	// Use cp -a to preserve permissions and symlinks
	cmd := exec.Command("cp", "-a", repoPath, destPath)
	if err := cmd.Run(); err != nil {
		return "", errors.Wrapf(err, "failed to copy staging to %s", destPath)
	}

	return destPath, nil
}

// CheckoutVersionInWorkspace checks out a specific version (tag, branch, or commit SHA)
// in a workspace copy of the repository. This is used when --module-version is specified
// to checkout a different version than what's in staging.
func CheckoutVersionInWorkspace(workspacePath, version string) error {
	if version == "" {
		return nil
	}

	cliprint.PrintStep(fmt.Sprintf("Checking out module version: %s", version))

	// Fetch all to ensure we have the version available (capture output)
	fetchCmd := exec.Command("git", "-C", workspacePath, "fetch", "--all", "--tags")
	var fetchStderr bytes.Buffer
	fetchCmd.Stderr = &fetchStderr
	if err := fetchCmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to fetch tags/branches: %s", fetchStderr.String())
	}

	// Checkout the version (capture output for error reporting)
	checkoutCmd := exec.Command("git", "-C", workspacePath, "checkout", version)
	var checkoutStderr bytes.Buffer
	checkoutCmd.Stderr = &checkoutStderr
	if err := checkoutCmd.Run(); err != nil {
		errMsg := checkoutStderr.String()
		if strings.Contains(errMsg, "did not match any") {
			return errors.Errorf("version '%s' not found. It may be a tag, branch, or commit SHA that doesn't exist", version)
		}
		return errors.Wrapf(err, "failed to checkout version '%s': %s", version, errMsg)
	}

	cliprint.PrintSuccess(fmt.Sprintf("Module version '%s' checked out", version))
	return nil
}

// CleanupWorkspaceCopy removes the copied repository from the workspace
func CleanupWorkspaceCopy(repoPath string) error {
	if repoPath == "" {
		return nil
	}

	// Safety check: don't delete the staging directory
	stagingPath, err := GetStagingRepoPath()
	if err != nil {
		return err
	}

	if repoPath == stagingPath {
		return errors.New("refusing to delete staging directory")
	}

	if !fileutil.IsDirExists(repoPath) {
		return nil // Already cleaned up
	}

	if err := os.RemoveAll(repoPath); err != nil {
		return errors.Wrapf(err, "failed to cleanup workspace copy at %s", repoPath)
	}

	return nil
}

// GetStagingInfo returns information about the current staging state
func GetStagingInfo() (exists bool, version string, repoPath string, err error) {
	exists, err = StagingExists()
	if err != nil {
		return false, "", "", err
	}

	if !exists {
		return false, "", "", nil
	}

	version, err = GetCurrentStagingVersion()
	if err != nil {
		return exists, "", "", err
	}

	repoPath, err = GetStagingRepoPath()
	if err != nil {
		return exists, version, "", err
	}

	return exists, version, repoPath, nil
}

// copyDir recursively copies a directory tree
// Note: This is kept as a fallback but we prefer cp -a for efficiency
func copyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate the destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
