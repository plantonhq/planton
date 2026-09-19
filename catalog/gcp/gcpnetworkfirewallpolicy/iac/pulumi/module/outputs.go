package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpPolicyName is the policy's name in GCP.
	OpPolicyName = "policy_name"
	// OpPolicyId is the policy's server-assigned numeric ID.
	OpPolicyId = "policy_id"
	// OpSelfLink is the policy's self-link URL.
	OpSelfLink = "self_link"
	// OpRegion is the policy's region, or empty for a global policy.
	OpRegion = "region"
	// OpRuleTupleCount is Google's complexity measure for the rule set.
	OpRuleTupleCount = "rule_tuple_count"
	// OpAssociationNames lists the association names in declaration order.
	OpAssociationNames = "association_names"
)
