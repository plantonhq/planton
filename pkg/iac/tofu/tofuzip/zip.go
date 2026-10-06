package tofuzip

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/version"
	"github.com/plantonhq/planton/internal/cli/workspace"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/downloads"
	"github.com/plantonhq/planton/pkg/fileutil"
)

const (
	// TerraformDirName is the base directory name for all Terraform-related files
	// All Terraform files are stored under ~/.planton/terraform/
	TerraformDirName = "terraform"

	// ModulesSubDir is the subdirectory for cached modules
	// Full path: ~/.planton/terraform/modules/{version}/
	ModulesSubDir = "modules"
)

// GetTerraformBaseDir returns the base directory for all Terraform-related files
// (~/.planton/terraform/)
func GetTerraformBaseDir() (string, error) {
	workspaceDir, err := workspace.GetWorkspaceDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get workspace directory")
	}
	return filepath.Join(workspaceDir, TerraformDirName), nil
}

// GetModuleCacheDir returns the path to the module cache directory
// (~/.planton/terraform/modules/{version}/)
func GetModuleCacheDir(releaseVersion string) (string, error) {
	terraformBaseDir, err := GetTerraformBaseDir()
	if err != nil {
		return "", err
	}

	// Normalize version for directory name
	versionDir := releaseVersion
	if versionDir == "" || versionDir == version.DefaultVersion {
		versionDir = "dev"
	}

	return filepath.Join(terraformBaseDir, ModulesSubDir, versionDir), nil
}

// GetModulePath returns the expected path for a cached module folder
// (~/.planton/terraform/modules/{version}/{kind}/)
func GetModulePath(kindName, releaseVersion string) (string, error) {
	cacheDir, err := GetModuleCacheDir(releaseVersion)
	if err != nil {
		return "", err
	}

	// Module folder name is lowercase kind name
	moduleFolderName := strings.ToLower(kindName)
	return filepath.Join(cacheDir, moduleFolderName), nil
}

// BuildDownloadURL constructs the Cloudflare R2 download URL for a Terraform module zip.
//
// The key is versionless (one live module set per kind); the release tag
// segment versions the artifact. Known skew edge: when releaseVersion is an
// OLDER tag whose lanes uploaded the pre-anatomy key shape, the download 404s
// and the caller falls back to staging (a git checkout of that tag), which
// self-corrects.
//
// Examples:
//
//	BuildDownloadURL("AwsEcsService", "v0.3.50")
//	  -> https://downloads.planton.ai/releases/v0.3.50/modules/terraform/awsecsservice/module.zip
func BuildDownloadURL(kindName, releaseVersion string) (string, error) {
	// Validate against the registry before composing: an unknown kind must
	// fail plainly here, not as a 404 the fallback path silently absorbs.
	if _, err := catalogkindreflect.KindVersionDir(kindName); err != nil {
		return "", errors.Wrapf(err, "cannot build the download URL for the %s terraform module", kindName)
	}
	return downloads.BuildTerraformDownloadURL(kindName, releaseVersion), nil
}

// IsModuleCached checks if a module is already cached and has .tf files
func IsModuleCached(kindName, releaseVersion string) (bool, error) {
	modulePath, err := GetModulePath(kindName, releaseVersion)
	if err != nil {
		return false, err
	}

	// Check if directory exists
	if !fileutil.IsDirExists(modulePath) {
		return false, nil
	}

	// Verify it has .tf files
	entries, err := os.ReadDir(modulePath)
	if err != nil {
		return false, errors.Wrapf(err, "failed to read module directory at %s", modulePath)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tf") {
			return true, nil
		}
	}

	return false, nil
}

// EnsureModule ensures the module for a kind is downloaded and cached.
// The releaseVersion can be:
// - CLI version like "v0.3.2" (uses main planton release)
// - Module version like "v0.3.2+terraform.awsecsservice.20260108.0" (uses kind-specific release)
// Returns the path to the module folder.
func EnsureModule(kindName, releaseVersion string) (string, error) {
	// Check if already cached
	cached, err := IsModuleCached(kindName, releaseVersion)
	if err != nil {
		return "", errors.Wrap(err, "failed to check module cache")
	}

	modulePath, err := GetModulePath(kindName, releaseVersion)
	if err != nil {
		return "", err
	}

	if cached {
		cliprint.PrintSuccess(fmt.Sprintf("Using cached module: %s", filepath.Base(modulePath)))
		return modulePath, nil
	}

	// Download the module
	cliprint.PrintStep(fmt.Sprintf("Downloading Terraform module for %s...", kindName))

	if err := DownloadAndExtractZip(kindName, releaseVersion); err != nil {
		return "", errors.Wrapf(err, "failed to download module for %s", kindName)
	}

	cliprint.PrintSuccess(fmt.Sprintf("Module downloaded: %s", filepath.Base(modulePath)))
	return modulePath, nil
}

// DownloadAndExtractZip downloads and extracts a kind's Terraform module zip from Cloudflare R2.
func DownloadAndExtractZip(kindName, releaseVersion string) error {
	// Ensure cache directory exists
	cacheDir, err := GetModuleCacheDir(releaseVersion)
	if err != nil {
		return err
	}

	if !fileutil.IsDirExists(cacheDir) {
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return errors.Wrapf(err, "failed to create cache directory %s", cacheDir)
		}
	}

	// Build download URL - the release version IS the tag
	downloadURL, err := BuildDownloadURL(kindName, releaseVersion)
	if err != nil {
		return err
	}

	cliprint.PrintInfo(fmt.Sprintf("Downloading from: %s", downloadURL))

	// Download the zip file
	resp, err := http.Get(downloadURL)
	if err != nil {
		return errors.Wrapf(err, "failed to download from %s", downloadURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("download failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	// Create a temporary file to store the zip
	tmpFile, err := os.CreateTemp("", "terraform-module-*.zip")
	if err != nil {
		return errors.Wrap(err, "failed to create temporary file")
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // Clean up temp file

	// Copy response to temp file
	written, err := io.Copy(tmpFile, resp.Body)
	if err != nil {
		tmpFile.Close()
		return errors.Wrap(err, "failed to download zip file")
	}
	tmpFile.Close()

	cliprint.PrintInfo(fmt.Sprintf("Downloaded %d bytes", written))

	// Get the module destination path
	modulePath, err := GetModulePath(kindName, releaseVersion)
	if err != nil {
		return err
	}

	// Create module directory
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		return errors.Wrapf(err, "failed to create module directory %s", modulePath)
	}

	// Extract zip to module directory
	if err := fileutil.ExtractZip(tmpPath, modulePath); err != nil {
		// Clean up partial extraction
		os.RemoveAll(modulePath)
		return errors.Wrap(err, "failed to extract zip file")
	}

	return nil
}

// GetCurrentCLIVersion returns the current CLI version, falling back to "dev" if not set
func GetCurrentCLIVersion() string {
	if version.Version == "" || version.Version == version.DefaultVersion {
		return "dev"
	}
	return version.Version
}

// IsDevVersion checks if the current CLI is a development version
func IsDevVersion() bool {
	return version.Version == "" || version.Version == version.DefaultVersion
}

// CanUseZipMode checks if we can use zip download mode.
// Returns false for dev builds (where zips don't exist in releases).
func CanUseZipMode() bool {
	return !IsDevVersion()
}
