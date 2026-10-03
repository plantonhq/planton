// Package provisionerannotationkeys defines the manifest annotation key that selects
// the IaC provisioner.
//
// Platform-behavior signals live in metadata.annotations, never metadata.labels:
// labels are derived into cloud-provider tags by planton IaC modules, so a platform
// key there would leak internal detail onto the user's real infra components.
package provisionerannotationkeys

const (
	// ProvisionerAnnotationKey specifies which IaC provisioner to use
	// Supported values: "pulumi", "tofu", "terraform" (case-insensitive)
	ProvisionerAnnotationKey = "planton.dev/provisioner"
)
