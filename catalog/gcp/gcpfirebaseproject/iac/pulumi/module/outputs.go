package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpProjectId is the GCP project Firebase is enabled on.
	OpProjectId = "project_id"
	// OpProjectNumber is the project's number -- the FCM sender id.
	OpProjectNumber = "project_number"
	// OpDisplayName is the project's display name as Firebase shows it.
	OpDisplayName = "display_name"
	// OpDatabaseUrl is the default Realtime Database URL (empty without one).
	OpDatabaseUrl = "database_url"
	// OpStorageBucket is the default Cloud Storage for Firebase bucket name
	// (empty without one).
	OpStorageBucket = "storage_bucket"
	// OpLocationId is the project's default GCP resource location (empty
	// until finalized).
	OpLocationId = "location_id"
)
