output "name" {
  description = "Full resource name of the membership (projects/{project}/locations/{location}/memberships/{id}) -- what scopes and per-cluster feature settings reference"
  value       = google_gke_hub_membership.this.name
}

output "membership_id" {
  description = "The membership's ID"
  value       = google_gke_hub_membership.this.membership_id
}

output "location" {
  description = "The membership's location"
  value       = google_gke_hub_membership.this.location
}
