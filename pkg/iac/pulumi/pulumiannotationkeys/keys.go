// Package pulumiannotationkeys defines the manifest annotation keys that carry
// Pulumi stack-location configuration.
//
// Platform-behavior signals live in metadata.annotations, never metadata.labels:
// labels are derived into cloud-provider tags by planton IaC modules, so a platform
// key there would leak internal detail onto the user's real infra components.
package pulumiannotationkeys

const (
	// StackFqdnAnnotationKey is the primary annotation that takes precedence over individual kinds
	// Format: "organization/project/stack"
	StackFqdnAnnotationKey = "pulumi.planton.dev/stack.fqdn"

	// OrganizationAnnotationKey is used when stack.fqdn is not present
	OrganizationAnnotationKey = "pulumi.planton.dev/organization"

	// ProjectAnnotationKey is used when stack.fqdn is not present
	ProjectAnnotationKey = "pulumi.planton.dev/project"

	// StackNameAnnotationKey is used when stack.fqdn is not present
	StackNameAnnotationKey = "pulumi.planton.dev/stack.name"

	// BackendUrlAnnotationKey pins the pulumi state backend for this manifest
	// (s3://..., gs://..., azblob://..., file://..., or a Pulumi Cloud URL).
	// Without it, pulumi uses whatever ambient `pulumi login` state the
	// machine holds — fine interactively, ambiguous in CI. Precedence follows
	// the tofu backend convention: CLI flag > this annotation > environment.
	BackendUrlAnnotationKey = "pulumi.planton.dev/backend.url"
)
