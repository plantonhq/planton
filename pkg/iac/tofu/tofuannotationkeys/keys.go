// Package tofuannotationkeys builds the manifest annotation keys that carry
// Terraform/OpenTofu backend configuration.
//
// Platform-behavior signals live in metadata.annotations, never metadata.labels:
// labels are derived into cloud-provider tags by planton IaC modules, so a platform
// key there would leak internal detail onto the user's real provider resources.
package tofuannotationkeys

import "fmt"

// BackendTypeAnnotationKey returns the backend type annotation key for the given provisioner.
// The provisioner should be "terraform" or "tofu".
// Example: BackendTypeAnnotationKey("terraform") returns "terraform.planton.dev/backend.type"
func BackendTypeAnnotationKey(provisioner string) string {
	return fmt.Sprintf("%s.planton.dev/backend.type", provisioner)
}

// BackendBucketAnnotationKey returns the backend bucket annotation key for the given provisioner.
// The provisioner should be "terraform" or "tofu".
// Example: BackendBucketAnnotationKey("terraform") returns "terraform.planton.dev/backend.bucket"
func BackendBucketAnnotationKey(provisioner string) string {
	return fmt.Sprintf("%s.planton.dev/backend.bucket", provisioner)
}

// BackendKeyAnnotationKey returns the backend key annotation key for the given provisioner.
// This is the state file path within the bucket.
// The provisioner should be "terraform" or "tofu".
// Example: BackendKeyAnnotationKey("terraform") returns "terraform.planton.dev/backend.key"
func BackendKeyAnnotationKey(provisioner string) string {
	return fmt.Sprintf("%s.planton.dev/backend.key", provisioner)
}

// BackendRegionAnnotationKey returns the backend region annotation key for the given provisioner.
// This is required for S3 backends.
// The provisioner should be "terraform" or "tofu".
// Example: BackendRegionAnnotationKey("terraform") returns "terraform.planton.dev/backend.region"
func BackendRegionAnnotationKey(provisioner string) string {
	return fmt.Sprintf("%s.planton.dev/backend.region", provisioner)
}

// BackendEndpointAnnotationKey returns the backend endpoint annotation key for the given provisioner.
// This is required for S3-compatible backends like Cloudflare R2 or MinIO.
// The provisioner should be "terraform" or "tofu".
// Example: BackendEndpointAnnotationKey("terraform") returns "terraform.planton.dev/backend.endpoint"
func BackendEndpointAnnotationKey(provisioner string) string {
	return fmt.Sprintf("%s.planton.dev/backend.endpoint", provisioner)
}
