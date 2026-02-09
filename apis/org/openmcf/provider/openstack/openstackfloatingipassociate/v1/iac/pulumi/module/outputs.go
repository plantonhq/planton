package module

const (
	// OpId is the exported stack output containing the Terraform resource ID.
	OpId = "id"
	// OpFloatingIp is the exported stack output containing the floating IP address.
	OpFloatingIp = "floating_ip"
	// OpPortId is the exported stack output containing the associated port UUID.
	OpPortId = "port_id"
	// OpFixedIp is the exported stack output containing the mapped fixed IP.
	// Computed by OpenStack if not explicitly specified.
	OpFixedIp = "fixed_ip"
	// OpRegion is the exported stack output containing the OpenStack region.
	OpRegion = "region"
)
