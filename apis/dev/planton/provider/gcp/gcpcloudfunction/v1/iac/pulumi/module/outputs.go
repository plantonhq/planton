package module

// Keys exported by the gcp_cloud_function Pulumi module.
// These map to the fields in GcpCloudFunctionStackOutputs proto.
const (
	OpFunctionId          = "function_id"           // Fully qualified resource name of the function
	OpFunctionUrl         = "function_url"          // HTTPS URL (HTTP triggers only)
	OpServiceAccountEmail = "service_account_email" // Runtime service account email
	OpState               = "state"                 // Current state of the function
	OpCloudRunServiceId   = "cloud_run_service_id"  // Cloud Run service ID
	OpEventarcTriggerId   = "eventarc_trigger_id"   // Eventarc trigger ID (event triggers only)
)
