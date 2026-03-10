package downloads

import (
	"fmt"
	"strings"
)

const (
	// BaseURL is the base URL for OpenMCF artifact downloads hosted on Cloudflare R2.
	// All non-CLI release artifacts (Pulumi binaries, Terraform modules, content zips)
	// are published here. CLI binaries remain on GitHub Releases via GoReleaser.
	BaseURL = "https://downloads.openmcf.org/releases"
)

// BuildPulumiDownloadURL constructs the R2 download URL for a Pulumi component binary.
//
// URL format: https://downloads.openmcf.org/releases/{version}/modules/pulumi/{component}_{platform}.gz
//
// Examples (on darwin/arm64):
//
//	BuildPulumiDownloadURL("AwsEcsService", "v0.3.50", "darwin_arm64")
//	  -> https://downloads.openmcf.org/releases/v0.3.50/modules/pulumi/awsecsservice_darwin_arm64.gz
func BuildPulumiDownloadURL(component, releaseVersion, platform string) string {
	artifact := fmt.Sprintf("%s_%s.gz", strings.ToLower(component), platform)
	return fmt.Sprintf("%s/%s/modules/pulumi/%s", BaseURL, releaseVersion, artifact)
}

// BuildTerraformDownloadURL constructs the R2 download URL for a Terraform module zip.
//
// URL format: https://downloads.openmcf.org/releases/{version}/modules/terraform/{component}.zip
//
// Examples:
//
//	BuildTerraformDownloadURL("AwsEcsService", "v0.3.50")
//	  -> https://downloads.openmcf.org/releases/v0.3.50/modules/terraform/awsecsservice.zip
func BuildTerraformDownloadURL(component, releaseVersion string) string {
	artifact := fmt.Sprintf("%s.zip", strings.ToLower(component))
	return fmt.Sprintf("%s/%s/modules/terraform/%s", BaseURL, releaseVersion, artifact)
}
