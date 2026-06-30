###############################################################################
# Stack Outputs (maps to GcpCloudFunctionStackOutputs proto)
###############################################################################

output "function_id" {
  description = "Fully qualified resource name of the deployed function"
  value       = google_cloudfunctions2_function.function.id
}

output "function_url" {
  description = "HTTPS URL of the function (only for HTTP triggers)"
  value       = local.is_http_trigger ? google_cloudfunctions2_function.function.service_config[0].uri : ""
}

output "service_account_email" {
  description = "Service account email that the function runs as"
  value = try(
    google_cloudfunctions2_function.function.service_config[0].service_account_email,
    ""
  )
}

output "state" {
  description = "Current state of the function"
  value       = google_cloudfunctions2_function.function.state
}

output "cloud_run_service_id" {
  description = "Cloud Run service ID (Gen 2 functions are deployed as Cloud Run services)"
  value       = google_cloudfunctions2_function.function.name
}

output "eventarc_trigger_id" {
  description = "Eventarc trigger ID (only for event-driven functions)"
  value = !local.is_http_trigger ? try(
    google_cloudfunctions2_function.function.event_trigger[0].trigger,
    ""
  ) : ""
}

