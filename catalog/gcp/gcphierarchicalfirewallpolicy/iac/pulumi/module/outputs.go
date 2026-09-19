package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpPolicyId is the policy's server-assigned numeric ID (Google's name).
	OpPolicyId = "policy_id"
	// OpShortName is the policy's user-facing short name.
	OpShortName = "short_name"
	// OpSelfLink is the policy's self-link URL.
	OpSelfLink = "self_link"
	// OpParent is the node the policy lives under.
	OpParent = "parent"
	// OpRuleTupleCount is Google's complexity measure for the rule set.
	OpRuleTupleCount = "rule_tuple_count"
	// OpAssociationNames lists the association names in declaration order.
	OpAssociationNames = "association_names"
)
