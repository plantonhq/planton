# Full resource name (projects/{project}/locations/{region}/workerPools/{name}).
output "name" {
  description = "Full resource name of the worker pool"
  value       = google_cloud_run_v2_worker_pool.main.id
}

output "worker_pool_name" {
  description = "Bare worker pool name in GCP"
  value       = google_cloud_run_v2_worker_pool.main.name
}

output "uid" {
  description = "Server-generated unique identifier for the worker pool"
  value       = google_cloud_run_v2_worker_pool.main.uid
}

output "location" {
  description = "Region the worker pool runs in"
  value       = google_cloud_run_v2_worker_pool.main.location
}

output "project_id" {
  description = "GCP project the worker pool lives in"
  value       = google_cloud_run_v2_worker_pool.main.project
}

output "latest_created_revision" {
  description = "Name of the most recently created revision"
  value       = google_cloud_run_v2_worker_pool.main.latest_created_revision
}

output "latest_ready_revision" {
  description = "Name of the newest revision serving instances"
  value       = google_cloud_run_v2_worker_pool.main.latest_ready_revision
}

output "observed_generation" {
  description = "Generation the controller has reconciled"
  value       = google_cloud_run_v2_worker_pool.main.observed_generation
}

output "etag" {
  description = "Opaque version token for optimistic concurrency"
  value       = google_cloud_run_v2_worker_pool.main.etag
}
