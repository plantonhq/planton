# Exactly one engine resource exists (chosen by engine_type), so every
# engine-level output reads the one that does through one(concat(...)).
output "name" {
  description = "Full resource name of the engine (projects/{project}/locations/{location}/collections/{collection_id}/engines/{engine_id})"
  value = one(concat(
    google_discovery_engine_search_engine.this[*].name,
    google_discovery_engine_chat_engine.this[*].name,
    google_discovery_engine_recommendation_engine.this[*].name,
  ))
}

output "engine_id" {
  description = "The engine's id"
  value       = local.created_engine_id
}

output "location" {
  description = "The engine's location (global, us, or eu)"
  value       = var.spec.location
}

output "collection_id" {
  description = "The collection the engine lives in"
  value       = local.collection_id
}

output "engine_type" {
  description = "Which app was built: SEARCH, CHAT, or RECOMMENDATION"
  value       = local.engine_type
}

output "serving_config_name" {
  description = "Full resource name of the engine's default serving config when configured (empty otherwise)"
  value       = var.spec.serving_config != null ? one(google_discovery_engine_serving_config.this[*].name) : ""
}

output "widget_config_name" {
  description = "Full resource name of the widget config when configured (empty otherwise)"
  value       = var.spec.widget_config != null ? one(google_discovery_engine_widget_config.this[*].name) : ""
}

# Google reports the agent under chat_engine_metadata, a list of one.
output "dialogflow_agent" {
  description = "The Dialogflow CX agent a CHAT engine answers through (empty on other engine types)"
  value       = local.is_chat ? try(one(google_discovery_engine_chat_engine.this[*].chat_engine_metadata[0].dialogflow_agent), "") : ""
}

# Manifest order, so the lists read the same on both engines.
output "control_names" {
  description = "Full resource names of the controls, in manifest order"
  value       = [for control in var.spec.controls : google_discovery_engine_control.this[control.control_id].name]
}

output "assistant_names" {
  description = "Full resource names of the assistants, in manifest order"
  value       = [for assistant in var.spec.assistants : google_discovery_engine_assistant.this[assistant.assistant_id].name]
}
