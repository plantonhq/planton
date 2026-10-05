package module

const (
	// OpAccountId is the exported output containing the Cloudflare
	// account the profile was applied to -- the profile is an account
	// singleton, so the account is its identity.
	OpAccountId = "account_id"

	// OpGatewayUniqueId is the exported output containing the
	// Gateway-side identifier Cloudflare assigns the profile.
	OpGatewayUniqueId = "gateway_unique_id"

	// OpPolicyId is the exported output containing the profile's
	// policy identifier as reported by the device policy API.
	OpPolicyId = "policy_id"
)
