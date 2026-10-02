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
