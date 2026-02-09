package module

const (
	// OpFloatingIpId is the exported stack output containing the floating IP UUID.
	OpFloatingIpId = "floating_ip_id"
	// OpAddress is the exported stack output containing the allocated floating IP address.
	// This is the primary output -- consumed by OpenStackFloatingIpAssociate as an FK target.
	OpAddress = "address"
	// OpFloatingNetworkId is the exported stack output containing the external network UUID.
	OpFloatingNetworkId = "floating_network_id"
	// OpPortId is the exported stack output containing the associated port UUID.
	// Empty if the floating IP is not associated (allocation-only mode).
	OpPortId = "port_id"
	// OpFixedIp is the exported stack output containing the associated fixed IP.
	// Empty if the floating IP is not associated.
	OpFixedIp = "fixed_ip"
	// OpRegion is the exported stack output containing the OpenStack region.
	OpRegion = "region"
)
