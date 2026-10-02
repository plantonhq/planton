output "project_id" {
  description = "The fleet host project's ID -- what every fleet child's project_id references"
  value       = google_gke_hub_fleet.this.project
}

output "name" {
  description = "Full resource name of the fleet (projects/{project}/locations/global/fleets/default)"
  value       = google_gke_hub_fleet.this.id
}

output "uid" {
  description = "Google's unique identifier for the fleet"
  value       = google_gke_hub_fleet.this.uid
}
