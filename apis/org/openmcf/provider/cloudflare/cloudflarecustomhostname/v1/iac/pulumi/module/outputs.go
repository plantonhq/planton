package module

const (
	// OpCustomHostnameId is the exported stack output containing the custom hostname id.
	OpCustomHostnameId = "custom_hostname_id"
	// OpStatus is the exported stack output containing the activation status.
	OpStatus = "status"
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
	// OpVerificationErrors are any verification errors reported by Cloudflare.
	OpVerificationErrors = "verification_errors"
	// OpCreatedAt is the creation timestamp.
	OpCreatedAt = "created_at"
)
