package module

const (
	// OpRuleId is the exported stack output containing the rule UUID.
	OpRuleId = "rule_id"
	// OpSecurityGroupId is the exported stack output containing the parent security group UUID.
	OpSecurityGroupId = "security_group_id"
	// OpDirection is the exported stack output containing the rule direction.
	OpDirection = "direction"
	// OpProtocol is the exported stack output containing the rule protocol.
	OpProtocol = "protocol"
	// OpPortRangeMin is the exported stack output containing the lower port bound.
	OpPortRangeMin = "port_range_min"
	// OpPortRangeMax is the exported stack output containing the upper port bound.
	OpPortRangeMax = "port_range_max"
	// OpRegion is the exported stack output containing the OpenStack region.
	OpRegion = "region"
)
