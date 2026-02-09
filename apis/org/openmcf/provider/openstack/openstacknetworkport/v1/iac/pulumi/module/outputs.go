package module

const (
	// OpPortId is the exported stack output containing the port UUID.
	// This is the primary FK target for downstream components.
	OpPortId = "port_id"
	// OpMacAddress is the exported stack output containing the port's MAC address.
	OpMacAddress = "mac_address"
	// OpAllFixedIps is the exported stack output containing all assigned IP addresses.
	OpAllFixedIps = "all_fixed_ips"
	// OpAllSecurityGroupIds is the exported stack output containing all applied SG UUIDs.
	OpAllSecurityGroupIds = "all_security_group_ids"
	// OpRegion is the exported stack output containing the OpenStack region.
	OpRegion = "region"
)
