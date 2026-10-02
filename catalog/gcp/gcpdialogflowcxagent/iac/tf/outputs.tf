output "name" {
  description = "Full resource name of the agent (projects/{project}/locations/{location}/agents/{agent_id}) -- what a chat engine's dialogflow_agent_to_link takes"
  value       = google_dialogflow_cx_agent.this.id
}

output "agent_id" {
  description = "The id Google assigned the agent"
  value       = google_dialogflow_cx_agent.this.name
}

output "location" {
  description = "The location the agent lives in"
  value       = google_dialogflow_cx_agent.this.location
}

output "start_flow" {
  description = "Full resource name of the agent's default start flow"
  value       = google_dialogflow_cx_agent.this.start_flow
}

# Manifest order, so the lists read the same on both engines.
output "webhook_names" {
  description = "Full resource names of the declared webhooks, in manifest order"
  value       = [for webhook in var.spec.webhooks : google_dialogflow_cx_webhook.this[webhook.display_name].id]
}

output "tool_names" {
  description = "Full resource names of the declared tools, in manifest order"
  value       = [for tool in var.spec.tools : google_dialogflow_cx_tool.this[tool.display_name].id]
}

output "tool_version_names" {
  description = "Full resource names of the declared tool versions, tool by tool in manifest order"
  value = flatten([
    for tool in var.spec.tools : [
      for version in tool.versions : google_dialogflow_cx_tool_version.this["${tool.display_name}/${version.display_name}"].id
    ]
  ])
}

output "version_names" {
  description = "Full resource names of the declared flow versions, in manifest order"
  value = [
    for version in var.spec.versions : google_dialogflow_cx_version.this["${version.flow_id != "" ? version.flow_id : local.start_flow_id}/${version.display_name}"].id
  ]
}

output "environment_names" {
  description = "Full resource names of the declared environments, in manifest order"
  value       = [for environment in var.spec.environments : google_dialogflow_cx_environment.this[environment.display_name].id]
}

output "generative_settings_names" {
  description = "Resource names of the declared generative settings ({agent}/generativeSettings?languageCode={language}), in manifest order"
  value       = [for settings in var.spec.generative_settings : google_dialogflow_cx_generative_settings.this[settings.language_code].id]
}
