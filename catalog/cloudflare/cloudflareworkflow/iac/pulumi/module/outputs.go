package module

const (
	// OpWorkflowName is the exported output containing the workflow's
	// name -- its identity within the account, and what Worker workflow
	// bindings reference.
	OpWorkflowName = "workflow_name"
	// OpVersionId is the exported output containing the workflow
	// version the registration produced.
	OpVersionId = "version_id"
)
