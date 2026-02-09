package module

const (
	// OpPortId is the exported stack output containing the auto-created port UUID.
	// This is also the Terraform resource ID for the router interface.
	OpPortId = "port_id"
	// OpRouterId is the exported stack output containing the router UUID.
	OpRouterId = "router_id"
	// OpSubnetId is the exported stack output containing the subnet UUID.
	OpSubnetId = "subnet_id"
	// OpRegion is the exported stack output containing the OpenStack region.
	OpRegion = "region"
)
