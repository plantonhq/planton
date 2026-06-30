output "job_id" {
  description = "The fully qualified job ID (projects/{project}/locations/{location}/jobs/{name})"
  value       = google_cloud_scheduler_job.this.id
}

output "job_name" {
  description = "The short job name"
  value       = google_cloud_scheduler_job.this.name
}

output "state" {
  description = "The current state of the job (ENABLED, PAUSED, DISABLED, UPDATE_FAILED)"
  value       = google_cloud_scheduler_job.this.state
}
