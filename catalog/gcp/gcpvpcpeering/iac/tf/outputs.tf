# The peering entry's name on this side's network -- the resolved value,
# from whichever of the two resources was created.
output "peering_name" {
  description = "The peering entry's name on this side's network"
  value       = local.is_create_form ? google_compute_network_peering.this[0].name : google_compute_network_peering_routes_config.this[0].peering
}

# This side's network, as given (a self link when referenced).
output "network" {
  description = "This side's network"
  value       = var.spec.network
}

# ACTIVE or INACTIVE as GCP reports it. Only the create form owns the peering
# entry and can read its state; the routes-config form exports an empty
# string so the outputs message has the same shape in either form.
output "state" {
  description = "The peering's state on this side (create form only)"
  value       = local.is_create_form ? google_compute_network_peering.this[0].state : ""
}

# GCP's explanation of the state (create form only).
output "state_details" {
  description = "GCP's explanation of the peering state (create form only)"
  value       = local.is_create_form ? google_compute_network_peering.this[0].state_details : ""
}
