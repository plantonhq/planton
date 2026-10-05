package module

const (
	// OpCertificatePackId is the exported output containing the pack id.
	// No status output: pack issuance is asynchronous (initializing ->
	// pending_validation -> active), and a point-in-time phase is never a
	// stable output -- it flips on the first refresh after the
	// transition and re-plans forever.
	OpCertificatePackId = "certificate_pack_id"
	// No primary_certificate output: the server populates it asynchronously
	// after the order (absent at create, "0" seconds later, then the real
	// certificate id as issuance progresses -- measured live 2026-08-29) --
	// a transitioning value is never a stable output.
	// OpZoneId is the exported output containing the Cloudflare zone ID
	// the pack was ordered in (a pack's API identity is zone_id + certificate_pack_id).
	OpZoneId = "zone_id"
)
