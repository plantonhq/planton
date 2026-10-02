output "name" {
  description = "Full resource name of the group (projects/{project}/locations/{location}/reservationGroups/{name}) -- what a reservation's reservation_group takes"
  value       = google_bigquery_reservation_group.this.id
}

output "reservation_group_name" {
  description = "The group's name"
  value       = google_bigquery_reservation_group.this.name
}

output "location" {
  description = "The group's location"
  value       = google_bigquery_reservation_group.this.location
}
