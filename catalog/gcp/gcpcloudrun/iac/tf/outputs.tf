output "url" {
  description = "Canonical serving URL of the Cloud Run service"
  value       = google_cloud_run_v2_service.main.uri
}

output "service_name" {
  description = "Name of the Cloud Run service as created in GCP"
  value       = google_cloud_run_v2_service.main.name
}

# The service resource keeps the revision that was ready when its plan was
# made, so after an update it would still name the previous revision. The
# service read back after the apply names the one now serving.
data "google_cloud_run_v2_service" "applied" {
  name     = google_cloud_run_v2_service.main.name
  location = google_cloud_run_v2_service.main.location
  project  = google_cloud_run_v2_service.main.project

  depends_on = [google_cloud_run_v2_service.main]
}

output "revision" {
  description = "Name of the latest ready revision"
  value       = data.google_cloud_run_v2_service.applied.latest_ready_revision
}

output "location" {
  description = "Region the service is deployed in"
  value       = google_cloud_run_v2_service.main.location
}

output "uid" {
  description = "Server-assigned unique identifier of the service"
  value       = google_cloud_run_v2_service.main.uid
}

output "urls" {
  description = "Every URL serving this service"
  value       = google_cloud_run_v2_service.main.urls
}

output "project_id" {
  description = "GCP project the service runs in -- with service_name and location, the coordinate the platform finds it by"
  value       = google_cloud_run_v2_service.main.project
}
