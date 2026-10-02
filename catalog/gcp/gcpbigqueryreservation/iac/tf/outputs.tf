output "name" {
  description = "Full resource name of the reservation (projects/{project}/locations/{location}/reservations/{reservation_name})"
  value       = google_bigquery_reservation.this.id
}

output "reservation_name" {
  description = "The reservation's name"
  value       = google_bigquery_reservation.this.name
}

output "location" {
  description = "The reservation's location"
  value       = google_bigquery_reservation.this.location
}

output "assignment_names" {
  description = "The assignments' full resource names, in the order they are declared"
  value       = [for k in local.assignment_keys : google_bigquery_reservation_assignment.this[k].id]
}
