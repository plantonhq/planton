output "name" {
  description = "Full resource name of the schedule (projects/{project}/locations/{location}/schedules/{schedule_id})"
  value       = google_colab_schedule.this.id
}

output "schedule_id" {
  description = "The id Google assigned the schedule"
  value       = google_colab_schedule.this.name
}

output "location" {
  description = "The region the schedule lives in"
  value       = google_colab_schedule.this.location
}

output "state" {
  description = "The schedule's state as Google reported it at the end of the deploy"
  value       = google_colab_schedule.this.state
}

output "next_run_time" {
  description = "When Google will next start a run, as of the end of the deploy"
  value       = google_colab_schedule.this.next_run_time
}

output "started_run_count" {
  description = "How many runs the schedule had started, as of the end of the deploy"
  value       = google_colab_schedule.this.started_run_count
}
