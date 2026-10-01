output "name" {
  description = "Full resource name of the notification config"
  value = one(concat(
    google_scc_v2_project_notification_config.this[*].name,
    google_scc_v2_folder_notification_config.this[*].name,
    google_scc_v2_organization_notification_config.this[*].name,
  ))
}

output "service_account" {
  description = "The Security Command Center service account that publishes to the topic, as a bare email (grant it roles/pubsub.publisher through a GcpPubSubTopicIamMember referencing service_account_member)"
  value       = local.service_account
}

# Composed exactly as the Pulumi module's exportOutputs does.
output "service_account_member" {
  description = "The publisher in IAM member form (serviceAccount:<email>) -- the member a GcpPubSubTopicIamMember grants roles/pubsub.publisher on the topic"
  value       = "serviceAccount:${local.service_account}"
}
