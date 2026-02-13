package module

const (
	// OpZoneName is the computed zone name ("{subdomain}.{domain}" or
	// just "{domain}" for root zones). This is the primary output for
	// downstream ScalewayDnsRecord references.
	OpZoneName = "zone_name"

	// OpNameServers is the list of nameservers assigned by Scaleway.
	// Users must configure these at their domain registrar for delegation.
	OpNameServers = "name_servers"

	// OpNameServersDefault is Scaleway's default nameserver list.
	OpNameServersDefault = "name_servers_default"

	// OpNameServersMaster is the master nameserver list.
	OpNameServersMaster = "name_servers_master"

	// OpStatus is the zone's current status (e.g., "active").
	OpStatus = "status"
)
