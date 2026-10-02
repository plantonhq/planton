package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpName is the group's resource name, groups/{group_id}.
	OpName = "name"
	// OpGroupEmail is the group's identity in IAM bindings.
	OpGroupEmail = "group_email"
	// OpMembershipCount is the number of memberships this manifest manages.
	OpMembershipCount = "membership_count"
)
