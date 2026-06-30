package upgrade

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// githubReleasesURL is the GitHub API endpoint for releases
	githubReleasesURL = "https://api.github.com/repos/plantonhq/planton/releases"
	// httpTimeout is the timeout for HTTP requests
	httpTimeout = 10 * time.Second
)

// GitHubRelease represents a GitHub release from the API
type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

// cliVersionInfo holds parsed CLI version information for comparison
type cliVersionInfo struct {
	Tag      string
	Major    int
	Minor    int
	Patch    int
	Date     int // YYYYMMDD as integer for comparison
	Revision int // The .N suffix
}

// cliVersionRegex matches CLI release tags like v0.3.15-cli.20260113.0
var cliVersionRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)-cli\.(\d{8})\.(\d+)$`)

// semverRegex matches any semver-like version (v0.3.17, 0.3.17, v0.3.15-cli.20260113.0, etc.)
var semverRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)`)

// parseCliVersion parses a CLI version tag into its components
func parseCliVersion(tag string) (*cliVersionInfo, error) {
	matches := cliVersionRegex.FindStringSubmatch(tag)
	if matches == nil {
		return nil, fmt.Errorf("not a valid CLI version tag")
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])
	date, _ := strconv.Atoi(matches[4])
	revision, _ := strconv.Atoi(matches[5])

	return &cliVersionInfo{
		Tag:      tag,
		Major:    major,
		Minor:    minor,
		Patch:    patch,
		Date:     date,
		Revision: revision,
	}, nil
}

// isGreaterThan returns true if v is greater than other
func (v *cliVersionInfo) isGreaterThan(other *cliVersionInfo) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	if v.Patch != other.Patch {
		return v.Patch > other.Patch
	}
	if v.Date != other.Date {
		return v.Date > other.Date
	}
	return v.Revision > other.Revision
}

// GetLatestVersion fetches the latest CLI version from GitHub releases
// It considers both:
// - CLI-specific releases with -cli.YYYYMMDD.N suffix (legacy format)
// - Plain semver releases like v0.3.16 (current format)
// It excludes Terraform and Pulumi module releases
// Returns the highest version found
func GetLatestVersion() (string, error) {
	releases, err := fetchReleases()
	if err != nil {
		return "", err
	}

	if len(releases) == 0 {
		return "", fmt.Errorf("no releases found")
	}

	// Track the highest version found
	// We compare by semver, and for same semver, CLI format with date/revision wins
	var highestTag string
	var highestSemver *semverInfo
	var highestCli *cliVersionInfo

	for _, release := range releases {
		// Skip drafts and pre-releases
		if release.Draft || release.Prerelease {
			continue
		}

		tag := release.TagName

		// Skip non-CLI releases (Terraform modules, Pulumi modules, etc.)
		if !isCliRelease(tag) {
			continue
		}

		// Try to parse as CLI version first (v0.3.15-cli.20260113.1)
		if cliVer, err := parseCliVersion(tag); err == nil {
			semver := &semverInfo{Major: cliVer.Major, Minor: cliVer.Minor, Patch: cliVer.Patch}

			if highestSemver == nil || semver.isGreaterThan(highestSemver) {
				// New highest semver
				highestSemver = semver
				highestCli = cliVer
				highestTag = tag
			} else if highestSemver.isEqual(semver) && highestCli != nil {
				// Same semver, compare CLI parts (date, revision)
				if cliVer.isGreaterThan(highestCli) {
					highestCli = cliVer
					highestTag = tag
				}
			}
			continue
		}

		// Try to parse as plain semver (v0.3.16)
		if semver, err := parseSemver(tag); err == nil {
			if highestSemver == nil || semver.isGreaterThan(highestSemver) {
				// New highest semver
				highestSemver = semver
				highestCli = nil // Plain semver, no CLI parts
				highestTag = tag
			}
			// If same semver and we already have a CLI version, keep the CLI version
			// (it has more specific versioning)
		}
	}

	if highestTag == "" {
		return "", fmt.Errorf("no valid CLI releases found")
	}

	return highestTag, nil
}

// isCliRelease checks if a release tag is a CLI release
// CLI releases are either:
// - Plain semver: v0.3.16 (unified releases that include CLI)
// - CLI format: v0.3.15-cli.20260113.1 (CLI-specific auto-releases)
//
// Non-CLI releases (excluded):
// - App releases: v0.3.16-app.20260113.0
// - Website releases: v0.3.16-website.20260113.0
// - Pulumi modules: v0.3.16-pulumi.{component}.20260113.0
// - Terraform modules: v0.3.16-terraform.{component}.20260113.0
func isCliRelease(tag string) bool {
	tagLower := strings.ToLower(tag)

	// Explicitly include CLI format
	if strings.Contains(tagLower, "-cli.") {
		return true
	}

	// Exclude all other component-specific releases
	excludePatterns := []string{
		"-app.",
		"-website.",
		"-pulumi.",
		"-terraform.",
	}

	for _, pattern := range excludePatterns {
		if strings.Contains(tagLower, pattern) {
			return false
		}
	}

	// Check if it's plain semver (v0.3.16) with no suffix
	if _, err := parseSemver(tag); err == nil {
		// Count hyphens - plain semver like "v0.3.16" has none
		// "v0.3.16-beta" has one but isn't a known component type
		parts := strings.SplitN(tag, "-", 2)
		if len(parts) == 1 {
			// Pure semver like v0.3.16 - this is a unified release
			return true
		}
		// Has a hyphen suffix but not a known component type
		// Could be something like v0.3.16-beta or v0.3.16-rc1
		// These are pre-releases that might include CLI, so include them
		// (GitHub API marks actual pre-releases with the Prerelease field)
		return false
	}

	return false
}

// semverInfo holds just the semver part of a version
type semverInfo struct {
	Major int
	Minor int
	Patch int
}

// parseSemver extracts the semver (major.minor.patch) from any version string
func parseSemver(version string) (*semverInfo, error) {
	matches := semverRegex.FindStringSubmatch(version)
	if matches == nil {
		return nil, fmt.Errorf("not a valid semver")
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	return &semverInfo{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

// isGreaterThan returns true if s is greater than other (semver comparison)
func (s *semverInfo) isGreaterThan(other *semverInfo) bool {
	if s.Major != other.Major {
		return s.Major > other.Major
	}
	if s.Minor != other.Minor {
		return s.Minor > other.Minor
	}
	return s.Patch > other.Patch
}

// isEqual returns true if s equals other (semver comparison)
func (s *semverInfo) isEqual(other *semverInfo) bool {
	return s.Major == other.Major && s.Minor == other.Minor && s.Patch == other.Patch
}

// CompareVersions returns true if latestVersion is newer than currentVersion
func CompareVersions(currentVersion, latestVersion string) bool {
	// If versions are the same, no upgrade needed
	if currentVersion == latestVersion {
		return false
	}

	// If current is "dev" or empty, always consider latest as newer
	if currentVersion == "dev" || currentVersion == "" {
		return true
	}

	// Try to parse both as CLI versions first (full comparison including date/revision)
	currentCli, errCurrent := parseCliVersion(currentVersion)
	latestCli, errLatest := parseCliVersion(latestVersion)

	// If both parse as CLI versions, compare them properly (semver + date + revision)
	if errCurrent == nil && errLatest == nil {
		return latestCli.isGreaterThan(currentCli)
	}

	// Fall back to semver-only comparison
	// This handles cases like current="v0.3.17" vs latest="v0.3.15-cli.20260113.1"
	currentSemver, errCurrentSemver := parseSemver(currentVersion)
	latestSemver, errLatestSemver := parseSemver(latestVersion)

	if errCurrentSemver == nil && errLatestSemver == nil {
		// If current semver is greater than or equal to latest semver, no upgrade needed
		// (e.g., v0.3.17 >= v0.3.15 means no upgrade)
		if currentSemver.isGreaterThan(latestSemver) || currentSemver.isEqual(latestSemver) {
			return false
		}
		return true
	}

	// Last resort: string comparison (should rarely happen)
	return currentVersion != latestVersion
}

// ValidateVersion checks if a specific version exists in GitHub releases
// It accepts versions in both formats:
// - CLI format: v0.3.15-cli.20260113.0
// - Plain semver: v0.3.16
// Returns the normalized tag or error if not found
func ValidateVersion(targetVersion string) (string, error) {
	// Normalize version - ensure it has 'v' prefix for matching
	normalizedTarget := targetVersion
	if !strings.HasPrefix(normalizedTarget, "v") {
		normalizedTarget = "v" + normalizedTarget
	}

	// Validate format - must be either CLI format or plain semver
	_, cliErr := parseCliVersion(normalizedTarget)
	_, semverErr := parseSemver(normalizedTarget)
	if cliErr != nil && semverErr != nil {
		return "", fmt.Errorf("invalid version format: %s (expected format: v0.3.16 or v0.3.15-cli.20260113.0)", targetVersion)
	}

	// Fetch releases from GitHub
	releases, err := fetchReleases()
	if err != nil {
		return "", err
	}

	// Look for exact match
	for _, release := range releases {
		if release.Draft || release.Prerelease {
			continue
		}

		// Check for exact match (case-insensitive)
		if strings.EqualFold(release.TagName, normalizedTarget) {
			return release.TagName, nil
		}
	}

	return "", fmt.Errorf("version %s not found in releases", targetVersion)
}

// fetchReleases fetches all releases from GitHub
func fetchReleases() ([]GitHubRelease, error) {
	client := &http.Client{
		Timeout: httpTimeout,
	}

	req, err := http.NewRequest("GET", githubReleasesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set Accept header for GitHub API
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch releases: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var releases []GitHubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("failed to parse releases: %w", err)
	}

	return releases, nil
}
