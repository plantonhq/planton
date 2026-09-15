package module

const (
	// OpCheckId is the uptime check UUID (the API identity, and the import id).
	OpCheckId = "check_id"

	// OpAlertIds is the map of composed alert-row UUIDs keyed identically to
	// the Terraform module's for_each key ("<row index>-<alert name>") -- the
	// second half of each row's "{check_id},{alert_id}" import id.
	OpAlertIds = "alert_ids"
)
