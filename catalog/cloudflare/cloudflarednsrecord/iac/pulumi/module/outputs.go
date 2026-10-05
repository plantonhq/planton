package module

const (
	// OpRecordId is the exported output containing the DNS record ID.
	OpRecordId = "record_id"
	// OpRecordName is the exported output containing the DNS record name.
	OpRecordName = "record_name"
	// OpRecordType is the exported output containing the record type.
	OpRecordType = "record_type"
	// OpProxied is the exported output indicating if the record is proxied.
	OpProxied = "proxied"
	// OpZoneId is the exported output containing the Cloudflare zone ID
	// the record lives in (a record's API identity is zone_id + record_id).
	OpZoneId = "zone_id"
)
