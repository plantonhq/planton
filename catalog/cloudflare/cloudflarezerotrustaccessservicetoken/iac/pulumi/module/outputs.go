package module

const (
	// OpServiceTokenId is the exported output containing the service
	// token's UUID (the API identity used for import and policy rules).
	OpServiceTokenId = "service_token_id"

	// OpClientId is the exported output containing the Client ID the
	// machine client presents in the CF-Access-Client-ID header.
	OpClientId = "client_id"

	// OpClientSecret is the exported output containing the Client Secret.
	// Cloudflare returns it only at creation and rotation -- the export is
	// marked secret and must be captured into a secret store at deploy time.
	OpClientSecret = "client_secret"

	// OpExpiresAt is the exported output containing the token's RFC3339
	// expiry.
	OpExpiresAt = "expires_at"
)
