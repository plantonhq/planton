package module

const (
	// OpCreatedAt is the exported output containing the creation timestamp.
	// No status output: fallback-origin deployment is asynchronous
	// (pending_deployment -> active), and a point-in-time phase is never a
	// stable output -- it flips on the first refresh after the
	// transition and re-plans forever.
	OpCreatedAt = "created_at"
	// OpUpdatedAt is the exported output containing the last-update timestamp.
	OpUpdatedAt = "updated_at"
	// No errors output: the server populates and clears the deployment-errors
	// list asynchronously after apply (the sibling custom hostname's
	// verification_errors measured live 2026-08-29) -- a transient diagnostic
	// is never a stable output.
	// OpZoneId is the exported output containing the Cloudflare zone ID
	// this singleton belongs to (the fallback origin has no resource id; its
	// API identity IS the zone).
	OpZoneId = "zone_id"
)
