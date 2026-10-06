package downloads

import (
	"fmt"
	"strings"
)

const (
	// BaseURL is the base URL for Planton artifact downloads hosted on Cloudflare R2.
	// All non-CLI release artifacts (Pulumi binaries, Terraform modules, content zips)
	// are published here. CLI binaries remain on GitHub Releases via GoReleaser.
	BaseURL = "https://downloads.planton.ai/releases"
)

// This package is pure URL grammar: every function composes the R2 key shapes
// that the release workflows upload (release.terraform-modules.yaml,
// release.pulumi-modules.yaml). The two sides MUST stay in lockstep — a shape
// change here or in CI without the matching change on the other side does not
// fail loudly; it turns every module fetch into a 404 that downstream code
// masks with a slow git-clone fallback. The unit tests in this package pin the
// shapes, and the release lanes probe one uploaded artifact per run, so drift
// surfaces in CI rather than in the field.
//
// Module artifacts are keyed per kind ({kind}/...): a kind has
// exactly one live module set regardless of how many API versions it serves
// (modules speak the hub version; older versions convert at the boundary), so
// the key carries no version segment. The release tag above the key versions
// the artifact, so shape stability holds across releases.

// BuildPulumiDownloadURL constructs the R2 download URL for a Pulumi module binary.
//
// URL format: https://downloads.planton.ai/releases/{version}/modules/pulumi/{kind}/{platform}.gz
// (windows artifacts carry the executable suffix: {platform}.exe.gz)
//
// Examples (on darwin/arm64):
//
//	BuildPulumiDownloadURL("AwsEcsService", "v0.3.50", "darwin_arm64")
//	  -> https://downloads.planton.ai/releases/v0.3.50/modules/pulumi/awsecsservice/darwin_arm64.gz
func BuildPulumiDownloadURL(kindDir, releaseVersion, platform string) string {
	// The release lane gzips "{kind}.exe" on windows, so the remote
	// artifact name carries ".exe.gz" there; local cache naming re-adds the
	// ".exe" independently (see pulumibinary.BuildBinaryName).
	suffix := ".gz"
	if strings.HasPrefix(platform, "windows") {
		suffix = ".exe.gz"
	}
	return fmt.Sprintf("%s/%s/modules/pulumi/%s/%s%s", BaseURL, releaseVersion, strings.ToLower(kindDir), platform, suffix)
}

// BuildPulumiSourceDownloadURL constructs the R2 download URL for a Pulumi
// kind's source zip — the module's Go source tree, as opposed to the
// compiled per-platform binaries served by BuildPulumiDownloadURL. Source is
// platform-independent, so the key carries no platform segment.
//
// URL format: https://downloads.planton.ai/releases/{version}/modules/pulumi/{kind}/source.zip
//
// Examples:
//
//	BuildPulumiSourceDownloadURL("AwsEcsService", "v0.3.50")
//	  -> https://downloads.planton.ai/releases/v0.3.50/modules/pulumi/awsecsservice/source.zip
func BuildPulumiSourceDownloadURL(kindDir, releaseVersion string) string {
	return fmt.Sprintf("%s/%s/modules/pulumi/%s/source.zip", BaseURL, releaseVersion, strings.ToLower(kindDir))
}

// BuildTerraformDownloadURL constructs the R2 download URL for a Terraform module zip.
//
// URL format: https://downloads.planton.ai/releases/{version}/modules/terraform/{kind}/module.zip
//
// Examples:
//
//	BuildTerraformDownloadURL("AwsEcsService", "v0.3.50")
//	  -> https://downloads.planton.ai/releases/v0.3.50/modules/terraform/awsecsservice/module.zip
func BuildTerraformDownloadURL(kindDir, releaseVersion string) string {
	return fmt.Sprintf("%s/%s/modules/terraform/%s/module.zip", BaseURL, releaseVersion, strings.ToLower(kindDir))
}

// BuildDefinitionsDownloadURL constructs the R2 download URL for a file in
// a definitions release (the skill/agent content packaged by
// pkg/skills/defspack: per-skill zips, agent instruction files, and the
// definitions-manifest.json that indexes them with checksums).
//
// URL format: https://downloads.planton.ai/releases/{version}/definitions/{file}
//
// Examples:
//
//	BuildDefinitionsDownloadURL("v0.4.0", "definitions-manifest.json")
//	  -> https://downloads.planton.ai/releases/v0.4.0/definitions/definitions-manifest.json
//	BuildDefinitionsDownloadURL("v0.4.0", "skill-planton.zip")
//	  -> https://downloads.planton.ai/releases/v0.4.0/definitions/skill-planton.zip
func BuildDefinitionsDownloadURL(releaseVersion, file string) string {
	return fmt.Sprintf("%s/%s/definitions/%s", BaseURL, releaseVersion, file)
}

// BuildDefinitionsExplodedFileURL constructs the R2 download URL for one
// file of a skill's exploded layout -- the same bytes as the skill
// archive's entry, published individually so reading surfaces fetch one
// file per click instead of the whole archive. The release's
// definitions-browse.json lists every path this shape serves.
//
// URL format: https://downloads.planton.ai/releases/{version}/definitions/exploded/{skillSlug}/{path}
//
// Examples:
//
//	BuildDefinitionsExplodedFileURL("v0.4.0", "planton", "references/chart-format.md")
//	  -> https://downloads.planton.ai/releases/v0.4.0/definitions/exploded/planton/references/chart-format.md
func BuildDefinitionsExplodedFileURL(releaseVersion, skillSlug, path string) string {
	return fmt.Sprintf("%s/%s/definitions/exploded/%s/%s", BaseURL, releaseVersion, skillSlug, path)
}
