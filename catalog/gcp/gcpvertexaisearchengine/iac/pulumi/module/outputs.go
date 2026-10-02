package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpName              = "name"
	OpEngineId          = "engine_id"
	OpLocation          = "location"
	OpCollectionId      = "collection_id"
	OpEngineType        = "engine_type"
	OpServingConfigName = "serving_config_name"
	OpWidgetConfigName  = "widget_config_name"
	OpDialogflowAgent   = "dialogflow_agent"
	OpControlNames      = "control_names"
	OpAssistantNames    = "assistant_names"
)
