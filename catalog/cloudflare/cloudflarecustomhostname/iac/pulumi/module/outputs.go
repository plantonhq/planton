package module

const (
	// OpCustomHostnameId is the exported output containing the custom hostname id.
	// No status output: hostname activation is asynchronous (pending ->
	// pending_validation -> active), and a point-in-time phase is never a
	// stable output -- it flips on the first refresh after the
	// transition and re-plans forever.
	OpCustomHostnameId = "custom_hostname_id"
	// OpOwnershipVerificationName is the DNS record name for ownership verification.
	OpOwnershipVerificationName = "ownership_verification_name"
	// OpOwnershipVerificationType is the DNS record type for ownership verification.
	OpOwnershipVerificationType = "ownership_verification_type"
	// OpOwnershipVerificationValue is the DNS record value for ownership verification.
	OpOwnershipVerificationValue = "ownership_verification_value"
	// OpOwnershipVerificationHttpUrl is the HTTP verification URL.
	OpOwnershipVerificationHttpUrl = "ownership_verification_http_url"
	// OpOwnershipVerificationHttpBody is the HTTP verification body.
	OpOwnershipVerificationHttpBody = "ownership_verification_http_body"
	// No verification_errors output: the server populates the list
	// asynchronously after apply ("zone is not active yet" measured appearing
	// seconds post-create) and clears it on activation -- a transient
	// diagnostic is never a stable output.
	// OpCreatedAt is the creation timestamp.
	OpCreatedAt = "created_at"
	// OpZoneId is the exported output containing the Cloudflare zone ID
	// the hostname was onboarded onto (API identity is zone_id + custom_hostname_id).
	OpZoneId = "zone_id"
)
