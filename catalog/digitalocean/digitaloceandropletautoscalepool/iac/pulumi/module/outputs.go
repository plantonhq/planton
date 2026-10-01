package module

const (
	// OpPoolId is the pool's UUID (its API identity and import id). The
	// pool's health is deliberately not exported: an apply-time status goes
	// stale the moment DigitalOcean changes it, so live health is read from
	// the API, never from stored outputs.
	OpPoolId = "pool_id"
)
