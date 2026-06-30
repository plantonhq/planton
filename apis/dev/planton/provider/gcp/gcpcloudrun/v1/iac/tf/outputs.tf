# outputs.tf - Stack outputs for GCP Cloud Run service

output "url" {
  description = "Public or internal URL of the Cloud Run service"
  value       = google_cloud_run_v2_service.main.uri
}

output "service_name" {
  description = "Name of the Cloud Run service"
  value       = google_cloud_run_v2_service.main.name
}

output "revision" {
  description = "Name of the latest deployed revision"
  value       = google_cloud_run_v2_service.main.latest_ready_revision
}
