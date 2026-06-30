package module

// Output key constants for Pulumi stack exports
// These constants define the keys used when exporting stack outputs
const (
	// OpId is the database resource ID
	OpId = "id"

	// OpName is the database name
	OpName = "name"

	// OpIsTransient indicates if the database is transient (cost optimization)
	OpIsTransient = "is_transient"

	// OpDataRetentionDays is the number of days for Time Travel retention
	OpDataRetentionDays = "data_retention_days"
)
