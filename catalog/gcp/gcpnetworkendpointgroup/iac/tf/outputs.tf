# Exactly one of the two group resources exists (see the count guards in
# main.tf), so each output selects whichever branch was created with
# one(concat(...)).

# The composition handle: a GcpBackendService backend names this as its
# group. A zonal link carries zones/{zone}, a global one says global.
output "self_link" {
  description = "Self-link URI of the network endpoint group (a backend service's backends[].group)"
  value       = one(concat(google_compute_network_endpoint_group.this[*].self_link, google_compute_global_network_endpoint_group.this[*].self_link))
}

output "neg_name" {
  description = "Name of the network endpoint group as it exists in GCP"
  value       = one(concat(google_compute_network_endpoint_group.this[*].name, google_compute_global_network_endpoint_group.this[*].name))
}

# The global collection exposes no server-generated numeric id through the
# provider; the zonal one does.
output "neg_id" {
  description = "Google's unique identifier for the group (empty for a global group)"
  value       = local.is_zonal ? tostring(one(google_compute_network_endpoint_group.this[*].generated_id)) : ""
}

output "zone" {
  description = "Zone of a zonal network endpoint group; empty for a global one"
  value       = local.is_zonal ? var.spec.zone : ""
}

# The declared membership: Google's own size attribute on the zonal group is
# read when the group is created, before the endpoint set is written, so it
# would report 0 here; the manifest's list is the membership on both scopes.
output "size" {
  description = "Number of endpoints declared in the group"
  value       = tostring(length(var.spec.endpoints))
}
