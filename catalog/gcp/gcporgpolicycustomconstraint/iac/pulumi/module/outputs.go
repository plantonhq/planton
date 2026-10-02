package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpName is the constraint's full resource name,
	// organizations/{org}/customConstraints/custom.{name}.
	OpName = "name"
	// OpConstraint is the custom.{name} handle a GcpOrgPolicy enforces.
	OpConstraint = "constraint"
	// OpUpdateTime is the RFC 3339 last-update timestamp.
	OpUpdateTime = "update_time"
)
