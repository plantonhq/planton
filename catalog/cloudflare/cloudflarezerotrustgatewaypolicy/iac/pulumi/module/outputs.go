package module

const (
	// OpPolicyId is the exported output containing the created Gateway
	// policy's UUID.
	OpPolicyId = "policy_id"

	// OpPrecedence is the exported output containing the policy's
	// evaluation precedence (assigned by Cloudflare when the spec leaves it
	// unset; lower runs earlier).
	OpPrecedence = "precedence"
)
