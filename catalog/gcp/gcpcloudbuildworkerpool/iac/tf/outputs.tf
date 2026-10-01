output "name" {
  description = "Full resource name of the pool (projects/{project}/locations/{location}/workerPools/{worker_pool_id})"
  value       = google_cloudbuild_worker_pool.this.id
}

output "worker_pool_id" {
  description = "The pool's ID"
  value       = google_cloudbuild_worker_pool.this.name
}

output "state" {
  description = "The pool's state (CREATING, RUNNING, UPDATING, DELETING, or DELETED)"
  value       = google_cloudbuild_worker_pool.this.state
}

output "uid" {
  description = "Google's unique identifier for the pool"
  value       = google_cloudbuild_worker_pool.this.uid
}
