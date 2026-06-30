package module

// Gateway API CRD manifest URLs for different channels and versions
const (
	// Base URL for Gateway API releases
	GatewayAPIReleaseBaseURL = "https://github.com/kubernetes-sigs/gateway-api/releases/download"

	// Standard channel manifest filename
	StandardInstallManifest = "standard-install.yaml"

	// Experimental channel manifest filename
	ExperimentalInstallManifest = "experimental-install.yaml"

	// Default Gateway API version
	DefaultVersion = "v1.2.1"
)

// GetManifestURL returns the download URL for Gateway API CRD manifests
func GetManifestURL(version string, experimental bool) string {
	manifest := StandardInstallManifest
	if experimental {
		manifest = ExperimentalInstallManifest
	}
	return GatewayAPIReleaseBaseURL + "/" + version + "/" + manifest
}
