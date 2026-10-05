package module

const (
	// OpJobId is the exported output containing the
	// Cloudflare-assigned numeric job ID, in string form.
	OpJobId = "job_id"

	// OpAccountId is the exported output containing the account the
	// job lives in (account-scoped jobs).
	OpAccountId = "account_id"

	// OpZoneId is the exported output containing the zone the job
	// lives in (zone-scoped jobs).
	OpZoneId = "zone_id"

	// OpOwnershipChallengeFilename is the exported output containing
	// the name of the challenge file Cloudflare dropped into the
	// destination (only when the challenge arm ran).
	OpOwnershipChallengeFilename = "ownership_challenge_filename"

	// OpOwnershipChallengeMessage is the exported output containing
	// the message accompanying the issued challenge.
	OpOwnershipChallengeMessage = "ownership_challenge_message"

	// OpOwnershipChallengeValid is the exported output reporting
	// whether Cloudflare found the destination valid when issuing the
	// challenge.
	OpOwnershipChallengeValid = "ownership_challenge_valid"
)
