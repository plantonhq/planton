package module

const (
	// OpZoneName    → the domain name (e.g. "example.com").
	OpZoneName = "zone_name"
	// OpZoneId → ID returned by DigitalOcean (identical to domain name).
	OpZoneId = "zone_id"
	// OpNameServers → list of authoritative NS records.
	OpNameServers = "name_servers"
)
