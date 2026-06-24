package module

const (
	// OpZoneId is the exported stack output containing the zone ID.
	OpZoneId = "zone_id"
	// OpNameservers is the exported stack output containing the assigned nameservers.
	OpNameservers = "nameservers"
	// OpStatus is the exported stack output containing the zone status.
	OpStatus = "status"
	// OpDnssecStatus is the exported DNSSEC status.
	OpDnssecStatus = "dnssec_status"
	// OpDnssecDs is the exported full DS record.
	OpDnssecDs = "dnssec_ds"
	// OpDnssecDigest is the exported DS digest.
	OpDnssecDigest = "dnssec_digest"
	// OpDnssecDigestType is the exported DS digest type code.
	OpDnssecDigestType = "dnssec_digest_type"
	// OpDnssecDigestAlgorithm is the exported DS digest algorithm.
	OpDnssecDigestAlgorithm = "dnssec_digest_algorithm"
	// OpDnssecAlgorithm is the exported DNSKEY algorithm code.
	OpDnssecAlgorithm = "dnssec_algorithm"
	// OpDnssecKeyTag is the exported DNSKEY key tag.
	OpDnssecKeyTag = "dnssec_key_tag"
	// OpDnssecPublicKey is the exported DNSKEY public key.
	OpDnssecPublicKey = "dnssec_public_key"
	// OpDnssecFlags is the exported DNSKEY flags.
	OpDnssecFlags = "dnssec_flags"
)
