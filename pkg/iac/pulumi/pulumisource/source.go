// Package pulumisource downloads and caches Pulumi kind module SOURCE
// from the release artifacts, the structural twin of tofuzip for the
// terraform half.
//
// The binaries beside this artifact (pulumibinary) serve execution; source
// serves the flows that need the module's code itself — ejecting an official
// module into a user-owned copy, and rendering module source without a git
// clone. Cache layout mirrors the sibling caches:
// ~/.planton/pulumi/sources/{version}/{kind}/.
package pulumisource

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
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/downloads"
	"github.com/plantonhq/planton/pkg/fileutil"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumibinary"
)

// SourcesSubDir is the subdirectory for cached module source trees.
// Full path: ~/.planton/pulumi/sources/{version}/
const SourcesSubDir = "sources"

// GetSourceCacheDir returns the path to the source cache directory
// (~/.planton/pulumi/sources/{version}/)
func GetSourceCacheDir(releaseVersion string) (string, error) {
	pulumiBaseDir, err := pulumibinary.GetPulumiBaseDir()
	if err != nil {
		return "", err
	}

	// Normalize version for directory name
	versionDir := releaseVersion
	if versionDir == "" || versionDir == version.DefaultVersion {
		versionDir = "dev"
	}

	return filepath.Join(pulumiBaseDir, SourcesSubDir, versionDir), nil
}

// GetSourcePath returns the expected path for a cached source folder
// (~/.planton/pulumi/sources/{version}/{kind}/)
func GetSourcePath(kindName, releaseVersion string) (string, error) {
	cacheDir, err := GetSourceCacheDir(releaseVersion)
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, strings.ToLower(kindName)), nil
}

// BuildDownloadURL constructs the Cloudflare R2 download URL for a Pulumi
// kind's source zip.
//
// The key is versionless (one live module set per kind); the release tag
// segment versions the artifact. Known skew edge: a tag whose release
// predates the pulumi source lane never uploaded a source.zip, so the
// download 404s and the caller falls back to a git checkout of that tag,
// which self-corrects.
func BuildDownloadURL(kindName, releaseVersion string) (string, error) {
	// Validate against the registry before composing: an unknown kind must
	// fail plainly here, not as a 404 the fallback path silently absorbs.
	if _, err := catalogkindreflect.KindVersionDir(kindName); err != nil {
		return "", errors.Wrapf(err, "cannot build the download URL for the %s pulumi module source", kindName)
	}
	return downloads.BuildPulumiSourceDownloadURL(kindName, releaseVersion), nil
}

// IsSourceCached checks if a kind's source is already cached and holds
// Go files.
func IsSourceCached(kindName, releaseVersion string) (bool, error) {
	sourcePath, err := GetSourcePath(kindName, releaseVersion)
	if err != nil {
		return false, err
	}

	if !fileutil.IsDirExists(sourcePath) {
		return false, nil
	}

	entries, err := os.ReadDir(sourcePath)
	if err != nil {
		return false, errors.Wrapf(err, "failed to read source directory at %s", sourcePath)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			return true, nil
		}
	}

	return false, nil
}

// EnsureSource ensures the source for a kind is downloaded and cached.
// Returns the path to the source folder.
func EnsureSource(kindName, releaseVersion string) (string, error) {
	cached, err := IsSourceCached(kindName, releaseVersion)
	if err != nil {
		return "", errors.Wrap(err, "failed to check source cache")
	}

	sourcePath, err := GetSourcePath(kindName, releaseVersion)
	if err != nil {
		return "", err
	}

	if cached {
		cliprint.PrintSuccess(fmt.Sprintf("Using cached module source: %s", filepath.Base(sourcePath)))
		return sourcePath, nil
	}

	cliprint.PrintStep(fmt.Sprintf("Downloading Pulumi module source for %s...", kindName))

	if err := downloadAndExtract(kindName, releaseVersion, sourcePath); err != nil {
		return "", errors.Wrapf(err, "failed to download module source for %s", kindName)
	}

	cliprint.PrintSuccess(fmt.Sprintf("Module source downloaded: %s", filepath.Base(sourcePath)))
	return sourcePath, nil
}

func downloadAndExtract(kindName, releaseVersion, sourcePath string) error {
	downloadURL, err := BuildDownloadURL(kindName, releaseVersion)
	if err != nil {
		return err
	}

	cliprint.PrintInfo(fmt.Sprintf("Downloading from: %s", downloadURL))

	resp, err := http.Get(downloadURL)
	if err != nil {
		return errors.Wrapf(err, "failed to download from %s", downloadURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("download failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	tmpFile, err := os.CreateTemp("", "pulumi-module-source-*.zip")
	if err != nil {
		return errors.Wrap(err, "failed to create temporary file")
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	written, err := io.Copy(tmpFile, resp.Body)
	if err != nil {
		tmpFile.Close()
		return errors.Wrap(err, "failed to download zip file")
	}
	tmpFile.Close()

	cliprint.PrintInfo(fmt.Sprintf("Downloaded %d bytes", written))

	if err := os.MkdirAll(sourcePath, 0755); err != nil {
		return errors.Wrapf(err, "failed to create source directory %s", sourcePath)
	}

	if err := fileutil.ExtractZip(tmpPath, sourcePath); err != nil {
		// Clean up partial extraction
		os.RemoveAll(sourcePath)
		return errors.Wrap(err, "failed to extract zip file")
	}

	return nil
}
