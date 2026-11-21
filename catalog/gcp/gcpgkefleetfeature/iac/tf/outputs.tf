output "name" {
  description = "Full resource name of the feature (projects/{project}/locations/{location}/features/{feature})"
  value       = google_gke_hub_feature.this.id
}

output "resource_state" {
  description = "The feature's lifecycle state in the fleet (ENABLING, ACTIVE, DISABLING, UPDATING, or SERVICE_UPDATING)"
  value       = try(google_gke_hub_feature.this.resource_state[0].state, "")
}

output "state_code" {
  description = "The feature's health as Google reports it (OK, WARNING, or ERROR); empty until assessed"
  value       = try(google_gke_hub_feature.this.state[0].state[0].code, "")
}
