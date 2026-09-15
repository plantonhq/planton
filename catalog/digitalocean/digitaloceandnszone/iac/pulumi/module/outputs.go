package module

const (
	// OpZoneName is the domain name (e.g. "example.com").
	OpZoneName = "zone_name"
	// OpZoneId is the zone's resource identifier — DigitalOcean addresses
	// domains by name, so this is the domain name itself.
	OpZoneId = "zone_id"
	// OpNameServers is DigitalOcean's fixed authoritative name server set.
	OpNameServers = "name_servers"
	// OpUrn is the uniform resource name of the domain.
	OpUrn = "urn"
	// OpRecordIds is the map of inline record ids keyed identically to the
	// Terraform module's for_each key (<record name>-<record index>-<value
	// index>) -- the import recipes resolve per-record ids through these
	// keys, so the two engines must agree on them.
	OpRecordIds = "record_ids"
)
