output "name" {
  description = "Full resource name of the commitment (projects/{project}/locations/{location}/capacityCommitments/{id})"
  value       = google_bigquery_capacity_commitment.this.name
}

output "state" {
  description = "The commitment's state (PENDING, ACTIVE, or FAILED)"
  value       = google_bigquery_capacity_commitment.this.state
}

output "commitment_start_time" {
  description = "When the current term started (ACTIVE commitments only)"
  value       = google_bigquery_capacity_commitment.this.commitment_start_time
}

output "commitment_end_time" {
  description = "When the current term ends (ACTIVE commitments only) -- the earliest a delete succeeds"
  value       = google_bigquery_capacity_commitment.this.commitment_end_time
}
