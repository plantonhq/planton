package pulumibinary

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
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
	// PulumiDirName is the base directory name for all Pulumi-related files
	// All Pulumi files are stored under ~/.planton/pulumi/
	PulumiDirName = "pulumi"

	// BinariesSubDir is the subdirectory for cached binaries
	// Full path: ~/.planton/pulumi/binaries/{version}/
	BinariesSubDir = "binaries"

	// WorkspacesSubDir is the subdirectory for Pulumi workspaces
	// Full path: ~/.planton/pulumi/workspaces/{stack-fqdn}/
	WorkspacesSubDir = "workspaces"

	// BinaryPrefix is the prefix for locally cached Pulumi module binaries.
	BinaryPrefix = "pulumi-"
)

// GetPlatformSuffix returns the platform suffix for binary names based on the current OS and architecture.
// Examples: "linux_amd64", "linux_arm64", "darwin_arm64", "darwin_amd64", "windows_amd64"
func GetPlatformSuffix() string {
	return fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
}

// GetPulumiBaseDir returns the base directory for all Pulumi-related files
// (~/.planton/pulumi/)
func GetPulumiBaseDir() (string, error) {
	workspaceDir, err := workspace.GetWorkspaceDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get workspace directory")
	}
	return filepath.Join(workspaceDir, PulumiDirName), nil
}

// GetBinaryCacheDir returns the path to the binary cache directory
// (~/.planton/pulumi/binaries/{version}/)
func GetBinaryCacheDir(releaseVersion string) (string, error) {
	pulumiBaseDir, err := GetPulumiBaseDir()
	if err != nil {
		return "", err
	}

	// Normalize version for directory name
	versionDir := releaseVersion
	if versionDir == "" || versionDir == version.DefaultVersion {
		versionDir = "dev"
	}

	return filepath.Join(pulumiBaseDir, BinariesSubDir, versionDir), nil
}

// GetBinaryPath returns the expected path for a cached binary
func GetBinaryPath(kindName, releaseVersion string) (string, error) {
	cacheDir, err := GetBinaryCacheDir(releaseVersion)
	if err != nil {
		return "", err
	}

	binaryName := BuildBinaryName(kindName)
	return filepath.Join(cacheDir, binaryName), nil
}

// BuildBinaryName constructs the platform-specific binary filename for a kind.
// The binary name includes the platform suffix based on the current OS and architecture.
// Examples:
//   - Linux:   "AwsEcsService" -> "pulumi-awsecsservice_linux_amd64"
//   - macOS:   "AwsEcsService" -> "pulumi-awsecsservice_darwin_arm64"
//   - Windows: "AwsEcsService" -> "pulumi-awsecsservice_windows_amd64.exe"
func BuildBinaryName(kindName string) string {
	baseName := BinaryPrefix + strings.ToLower(kindName)
	suffix := GetPlatformSuffix()

	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s_%s.exe", baseName, suffix)
	}
	return fmt.Sprintf("%s_%s", baseName, suffix)
}

// BuildDownloadURL constructs the Cloudflare R2 download URL for a platform-specific kind binary.
//
// The key is versionless (one live module set per kind); the release tag
// segment versions the artifact. Known skew edge: when releaseVersion is an
// OLDER tag whose lanes uploaded the pre-anatomy key shape, the download 404s
// and the caller falls back to staging (a git checkout of that tag), which
// self-corrects.
//
// Examples (on darwin/arm64):
//
//	BuildDownloadURL("AwsEcsService", "v0.3.50")
//	  -> https://downloads.planton.ai/releases/v0.3.50/modules/pulumi/awsecsservice/darwin_arm64.gz
func BuildDownloadURL(kindName, releaseVersion string) (string, error) {
	// Validate against the registry before composing: an unknown kind must
	// fail plainly here, not as a 404 the fallback path silently absorbs.
	if _, err := catalogkindreflect.KindVersionDir(kindName); err != nil {
		return "", errors.Wrapf(err, "cannot build the download URL for the %s pulumi module", kindName)
	}
	return downloads.BuildPulumiDownloadURL(kindName, releaseVersion, GetPlatformSuffix()), nil
}

// IsBinaryCached checks if a binary is already cached
func IsBinaryCached(kindName, releaseVersion string) (bool, error) {
	binaryPath, err := GetBinaryPath(kindName, releaseVersion)
	if err != nil {
		return false, err
	}

	exists, err := fileutil.IsExists(binaryPath)
	if err != nil {
		return false, errors.Wrapf(err, "failed to check if binary exists at %s", binaryPath)
	}

	if !exists {
		return false, nil
	}

	// Verify it's executable
	info, err := os.Stat(binaryPath)
	if err != nil {
		return false, errors.Wrapf(err, "failed to stat binary at %s", binaryPath)
	}

	// Check if file has execute permission
	return info.Mode()&0111 != 0, nil
}

// EnsureBinary ensures the binary for a kind is downloaded and cached.
// The releaseVersion can be:
// - CLI version like "v0.3.2" (uses main planton release)
// - Module version like "v0.3.1-pulumi-awsecsservice-20260107.01" (uses kind-specific release)
// Returns the path to the binary.
func EnsureBinary(kindName, releaseVersion string) (string, error) {
	// Check if already cached
	cached, err := IsBinaryCached(kindName, releaseVersion)
	if err != nil {
		return "", errors.Wrap(err, "failed to check binary cache")
	}

	binaryPath, err := GetBinaryPath(kindName, releaseVersion)
	if err != nil {
		return "", err
	}

	if cached {
		cliprint.PrintSuccess(fmt.Sprintf("Using cached module: %s", kindName))
		return binaryPath, nil
	}

	// Download the binary
	cliprint.PrintStep(fmt.Sprintf("Downloading Pulumi module: %s...", kindName))

	if err := DownloadBinary(kindName, releaseVersion); err != nil {
		return "", errors.Wrapf(err, "failed to download binary for %s", kindName)
	}

	cliprint.PrintSuccess(fmt.Sprintf("Module downloaded: %s", kindName))
	return binaryPath, nil
}

// DownloadBinary downloads and extracts a platform-specific kind binary from Cloudflare R2.
// The binary downloaded is platform-specific based on runtime.GOOS and runtime.GOARCH.
func DownloadBinary(kindName, releaseVersion string) error {
	// Ensure cache directory exists
	cacheDir, err := GetBinaryCacheDir(releaseVersion)
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

	// Download the gzipped binary
	resp, err := http.Get(downloadURL)
	if err != nil {
		return errors.Wrapf(err, "failed to download from %s", downloadURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("download failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	// Create gzip reader
	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to create gzip reader")
	}
	defer gzReader.Close()

	// Write to destination
	binaryPath, err := GetBinaryPath(kindName, releaseVersion)
	if err != nil {
		return err
	}

	outFile, err := os.OpenFile(binaryPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return errors.Wrapf(err, "failed to create binary file at %s", binaryPath)
	}
	defer outFile.Close()

	written, err := io.Copy(outFile, gzReader)
	if err != nil {
		// Clean up partial file
		os.Remove(binaryPath)
		return errors.Wrap(err, "failed to extract binary")
	}

	cliprint.PrintInfo(fmt.Sprintf("Extracted %d bytes", written))

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
