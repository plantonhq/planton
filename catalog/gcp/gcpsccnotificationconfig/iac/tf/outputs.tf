output "name" {
  description = "Full resource name of the notification config"
  value = one(concat(
    google_scc_v2_project_notification_config.this[*].name,
    google_scc_v2_folder_notification_config.this[*].name,
    google_scc_v2_organization_notification_config.this[*].name,
  ))
}

output "service_account" {
  description = "The Security Command Center service account that publishes to the topic (grant it roles/pubsub.publisher there)"
  value = one(concat(
    google_scc_v2_project_notification_config.this[*].service_account,
    google_scc_v2_folder_notification_config.this[*].service_account,
    google_scc_v2_organization_notification_config.this[*].service_account,
  ))
}
