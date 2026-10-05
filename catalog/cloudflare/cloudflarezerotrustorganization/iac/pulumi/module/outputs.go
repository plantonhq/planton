package module

const (
	// OpAuthDomain is the exported output containing the team domain
	// (without the .cloudflareaccess.com suffix).
	OpAuthDomain = "auth_domain"

	// OpAccountId is the exported output containing the account the
	// organization was applied to (empty for a zone-scoped organization) --
	// the singleton's identity for the harness and import recipes.
	OpAccountId = "account_id"
)
