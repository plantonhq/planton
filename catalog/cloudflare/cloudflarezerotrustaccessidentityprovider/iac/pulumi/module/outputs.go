package module

const (
	// OpIdentityProviderId is the exported output containing the identity
	// provider's UUID (what Access policy rules reference).
	OpIdentityProviderId = "identity_provider_id"

	// OpScimBaseUrl is the exported output containing the base URL of
	// Cloudflare's SCIM v2.0 endpoint for this provider (present when SCIM is
	// enabled).
	OpScimBaseUrl = "scim_base_url"

	// OpScimSecret is the exported output containing the SCIM bearer
	// token. Cloudflare mints it once when SCIM is first enabled and redacts it
	// on later reads -- the export is marked secret and must be captured into a
	// secret store at deploy time.
	OpScimSecret = "scim_secret"
)
