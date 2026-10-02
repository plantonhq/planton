package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpName                    = "name"
	OpAgentId                 = "agent_id"
	OpLocation                = "location"
	OpStartFlow               = "start_flow"
	OpWebhookNames            = "webhook_names"
	OpToolNames               = "tool_names"
	OpToolVersionNames        = "tool_version_names"
	OpVersionNames            = "version_names"
	OpEnvironmentNames        = "environment_names"
	OpGenerativeSettingsNames = "generative_settings_names"
)
