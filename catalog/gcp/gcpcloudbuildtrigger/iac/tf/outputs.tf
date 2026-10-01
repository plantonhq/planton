output "id" {
  description = "The trigger's full resource ID (projects/{project}/locations/{location}/triggers/{trigger_id}; projects/{project}/triggers/{trigger_id} for a global trigger)"
  value       = google_cloudbuild_trigger.this.id
}

output "trigger_id" {
  description = "Google's generated unique ID for the trigger"
  value       = google_cloudbuild_trigger.this.trigger_id
}

output "name" {
  description = "The trigger's name"
  value       = google_cloudbuild_trigger.this.name
}
