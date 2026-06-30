package pulumilabels

const (
	// StackFqdnLabelKey is the primary label that takes precedence over individual components
	// Format: "organization/project/stack"
	StackFqdnLabelKey = "pulumi.planton.dev/stack.fqdn"

	// OrganizationLabelKey is used when stack.fqdn is not present
	OrganizationLabelKey = "pulumi.planton.dev/organization"

	// ProjectLabelKey is used when stack.fqdn is not present
	ProjectLabelKey = "pulumi.planton.dev/project"

	// StackNameLabelKey is used when stack.fqdn is not present
	StackNameLabelKey = "pulumi.planton.dev/stack.name"
)
