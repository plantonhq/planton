package tofulabels

import "fmt"

// BackendTypeLabelKey returns the backend type label key for the given provisioner.
// The provisioner should be "terraform" or "tofu".
// Example: BackendTypeLabelKey("terraform") returns "terraform.planton.dev/backend.type"
func BackendTypeLabelKey(provisioner string) string {
	return fmt.Sprintf("%s.planton.dev/backend.type", provisioner)
}

// BackendBucketLabelKey returns the backend bucket label key for the given provisioner.
// The provisioner should be "terraform" or "tofu".
// Example: BackendBucketLabelKey("terraform") returns "terraform.planton.dev/backend.bucket"
func BackendBucketLabelKey(provisioner string) string {
	return fmt.Sprintf("%s.planton.dev/backend.bucket", provisioner)
}

// BackendKeyLabelKey returns the backend key label key for the given provisioner.
// This is the state file path within the bucket.
// The provisioner should be "terraform" or "tofu".
// Example: BackendKeyLabelKey("terraform") returns "terraform.planton.dev/backend.key"
func BackendKeyLabelKey(provisioner string) string {
	return fmt.Sprintf("%s.planton.dev/backend.key", provisioner)
}

// BackendRegionLabelKey returns the backend region label key for the given provisioner.
// This is required for S3 backends.
// The provisioner should be "terraform" or "tofu".
// Example: BackendRegionLabelKey("terraform") returns "terraform.planton.dev/backend.region"
func BackendRegionLabelKey(provisioner string) string {
	return fmt.Sprintf("%s.planton.dev/backend.region", provisioner)
}

// BackendEndpointLabelKey returns the backend endpoint label key for the given provisioner.
// This is required for S3-compatible backends like Cloudflare R2 or MinIO.
// The provisioner should be "terraform" or "tofu".
// Example: BackendEndpointLabelKey("terraform") returns "terraform.planton.dev/backend.endpoint"
func BackendEndpointLabelKey(provisioner string) string {
	return fmt.Sprintf("%s.planton.dev/backend.endpoint", provisioner)
}

// Legacy constants for backward compatibility.
// These are kept to ensure existing manifests using terraform.* labels
// continue to work regardless of the provisioner being used.
const (
	// LegacyBackendTypeLabelKey is the legacy backend type label (terraform prefix)
	LegacyBackendTypeLabelKey = "terraform.planton.dev/backend.type"

	// LegacyBackendBucketLabelKey is the legacy backend bucket label (terraform prefix)
	LegacyBackendBucketLabelKey = "terraform.planton.dev/backend.bucket"

	// LegacyBackendKeyLabelKey is the legacy backend key label (terraform prefix)
	LegacyBackendKeyLabelKey = "terraform.planton.dev/backend.key"

	// LegacyBackendObjectLabelKey is the deprecated backend object label (terraform prefix)
	// Kept for backward compatibility - prefer backend.key
	LegacyBackendObjectLabelKey = "terraform.planton.dev/backend.object"

	// LegacyBackendRegionLabelKey is the legacy backend region label (terraform prefix)
	LegacyBackendRegionLabelKey = "terraform.planton.dev/backend.region"

	// LegacyBackendEndpointLabelKey is the legacy backend endpoint label (terraform prefix)
	// Used for S3-compatible backends like Cloudflare R2 or MinIO
	LegacyBackendEndpointLabelKey = "terraform.planton.dev/backend.endpoint"
)
