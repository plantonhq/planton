output "name" {
  description = "Full resource name of the connection (projects/{project}/locations/{location}/connections/{connection_id})"
  value       = google_cloudbuildv2_connection.this.id
}

output "connection_id" {
  description = "The connection's ID"
  value       = google_cloudbuildv2_connection.this.name
}

output "installation_stage" {
  description = "The current installation step (PENDING_CREATE_APP, PENDING_USER_OAUTH, PENDING_INSTALL_APP, or COMPLETE)"
  value       = try(google_cloudbuildv2_connection.this.installation_state[0].stage, "")
}

output "installation_action_uri" {
  description = "The link a person follows to finish the installation; empty once complete"
  value       = try(google_cloudbuildv2_connection.this.installation_state[0].action_uri, "")
}
