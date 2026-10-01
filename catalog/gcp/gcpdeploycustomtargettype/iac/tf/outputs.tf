output "name" {
  description = "Full resource name of the custom target type (projects/{project}/locations/{location}/customTargetTypes/{custom_target_type_id})"
  value       = google_clouddeploy_custom_target_type.this.id
}

output "custom_target_type_id" {
  description = "The custom target type's ID"
  value       = google_clouddeploy_custom_target_type.this.name
}

output "uid" {
  description = "Google's unique identifier for the custom target type"
  value       = google_clouddeploy_custom_target_type.this.uid
}
