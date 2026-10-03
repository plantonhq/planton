package module

const (
	// OpRuleId is the exported output containing the created rule's ID.
	OpRuleId = "rule_id"
	// OpZoneId is the exported output containing the zone the rule was
	// created on (empty for account-wide rules).
	OpZoneId = "zone_id"
	// OpAccountId is the exported output containing the account the rule
	// was created on (empty for zone-scoped rules).
	OpAccountId = "account_id"
)
