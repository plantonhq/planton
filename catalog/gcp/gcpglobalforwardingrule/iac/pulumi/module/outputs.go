package module

// Output keys must match the field names in outputs.proto — the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpIpAddress           = "ip_address"
	OpSelfLink            = "self_link"
	OpForwardingRuleName  = "forwarding_rule_name"
	OpForwardingRuleId    = "forwarding_rule_id"
	OpPscConnectionId     = "psc_connection_id"
	OpPscConnectionStatus = "psc_connection_status"
	OpRegion              = "region"
	OpServiceName         = "service_name"
)
