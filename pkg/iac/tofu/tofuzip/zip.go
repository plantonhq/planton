package tofuzip

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/internal/cli/cliprint"
	"github.com/plantonhq/openmcf/internal/cli/version"
	"github.com/plantonhq/openmcf/internal/cli/workspace"
	"github.com/plantonhq/openmcf/pkg/downloads"
	"github.com/plantonhq/openmcf/pkg/fileutil"
)

const (
	// TerraformDirName is the base directory name for all Terraform-related files
	// All Terraform files are stored under ~/.openmcf/terraform/
	TerraformDirName = "terraform"

	// ModulesSubDir is the subdirectory for cached modules
	// Full path: ~/.openmcf/terraform/modules/{version}/
	ModulesSubDir = "modules"
)

// GetTerraformBaseDir returns the base directory for all Terraform-related files
// (~/.openmcf/terraform/)
func GetTerraformBaseDir() (string, error) {
	workspaceDir, err := workspace.GetWorkspaceDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get workspace directory")
	}
	return filepath.Join(workspaceDir, TerraformDirName), nil
}

// GetModuleCacheDir returns the path to the module cache directory
// (~/.openmcf/terraform/modules/{version}/)
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
// (~/.openmcf/terraform/modules/{version}/{component}/)
func GetModulePath(componentName, releaseVersion string) (string, error) {
	cacheDir, err := GetModuleCacheDir(releaseVersion)
	if err != nil {
		return "", err
	}

	// Module folder name is lowercase component name
	moduleFolderName := strings.ToLower(componentName)
	return filepath.Join(cacheDir, moduleFolderName), nil
}

// BuildDownloadURL constructs the Cloudflare R2 download URL for a Terraform module zip.
//
// Examples:
//
//	BuildDownloadURL("AwsEcsService", "v0.3.50")
//	  -> https://downloads.openmcf.org/releases/v0.3.50/modules/terraform/awsecsservice.zip
func BuildDownloadURL(componentName, releaseVersion string) string {
	return downloads.BuildTerraformDownloadURL(componentName, releaseVersion)
}

// IsModuleCached checks if a module is already cached and has .tf files
func IsModuleCached(componentName, releaseVersion string) (bool, error) {
	modulePath, err := GetModulePath(componentName, releaseVersion)
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

// EnsureModule ensures the module for a component is downloaded and cached.
// The releaseVersion can be:
// - CLI version like "v0.3.2" (uses main openmcf release)
// - Module version like "v0.3.2+terraform.awsecsservice.20260108.0" (uses component-specific release)
// Returns the path to the module folder.
func EnsureModule(componentName, releaseVersion string) (string, error) {
	// Check if already cached
	cached, err := IsModuleCached(componentName, releaseVersion)
	if err != nil {
		return "", errors.Wrap(err, "failed to check module cache")
	}

	modulePath, err := GetModulePath(componentName, releaseVersion)
	if err != nil {
		return "", err
	}

	if cached {
		cliprint.PrintSuccess(fmt.Sprintf("Using cached module: %s", filepath.Base(modulePath)))
		return modulePath, nil
	}

	// Download the module
	cliprint.PrintStep(fmt.Sprintf("Downloading Terraform module for %s...", componentName))

	if err := DownloadAndExtractZip(componentName, releaseVersion); err != nil {
		return "", errors.Wrapf(err, "failed to download module for %s", componentName)
	}

	cliprint.PrintSuccess(fmt.Sprintf("Module downloaded: %s", filepath.Base(modulePath)))
	return modulePath, nil
}

// DownloadAndExtractZip downloads and extracts a component's Terraform module zip from Cloudflare R2.
func DownloadAndExtractZip(componentName, releaseVersion string) error {
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
	downloadURL := BuildDownloadURL(componentName, releaseVersion)

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
	modulePath, err := GetModulePath(componentName, releaseVersion)
	if err != nil {
		return err
	}

	// Create module directory
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		return errors.Wrapf(err, "failed to create module directory %s", modulePath)
	}

	// Extract zip to module directory
	if err := extractZip(tmpPath, modulePath); err != nil {
		// Clean up partial extraction
		os.RemoveAll(modulePath)
		return errors.Wrap(err, "failed to extract zip file")
	}

	return nil
}

// extractZip extracts a zip file to the destination directory
func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return errors.Wrapf(err, "failed to open zip file %s", zipPath)
	}
	defer r.Close()

	for _, f := range r.File {
		// Construct the destination path
		destPath := filepath.Join(destDir, f.Name)

		// Check for zip slip vulnerability
		if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return errors.Errorf("illegal file path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			// Create directory
			if err := os.MkdirAll(destPath, f.Mode()); err != nil {
				return errors.Wrapf(err, "failed to create directory %s", destPath)
			}
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return errors.Wrapf(err, "failed to create parent directory for %s", destPath)
		}

		// Extract file
		if err := extractFile(f, destPath); err != nil {
			return err
		}
	}

	return nil
}

// extractFile extracts a single file from a zip archive
func extractFile(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return errors.Wrapf(err, "failed to open file in zip: %s", f.Name)
	}
	defer rc.Close()

	outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return errors.Wrapf(err, "failed to create file %s", destPath)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, rc); err != nil {
		return errors.Wrapf(err, "failed to write file %s", destPath)
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
