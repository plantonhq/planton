# Exactly one of the two resource families exists (see the count guards in
# main.tf), so each output selects whichever branch was created with
# one(concat(...)).

# The policy's name in GCP -- what gcloud's describe takes.
output "policy_name" {
  description = "Name of the network firewall policy"
  value       = one(concat(google_compute_network_firewall_policy.this[*].name, google_compute_region_network_firewall_policy.this[*].name))
}

# The server-assigned numeric ID.
output "policy_id" {
  description = "The policy's server-assigned numeric ID"
  value       = one(concat(google_compute_network_firewall_policy.this[*].network_firewall_policy_id, google_compute_region_network_firewall_policy.this[*].region_network_firewall_policy_id))
}

output "self_link" {
  description = "Self-link URL of the policy"
  value       = one(concat(google_compute_network_firewall_policy.this[*].self_link, google_compute_region_network_firewall_policy.this[*].self_link))
}

# Region of a regional policy; empty for a global one, so a consumer knows
# which resource family the policy is.
output "region" {
  description = "Region of the policy (empty for global)"
  value       = local.is_regional ? var.spec.region : ""
}

# Google's complexity measure for the whole rule set; a project has a quota
# of tuples per policy.
output "rule_tuple_count" {
  description = "Total rule tuple count of the policy"
  value       = one(concat(google_compute_network_firewall_policy.this[*].rule_tuple_count, google_compute_region_network_firewall_policy.this[*].rule_tuple_count))
}

# Declaration-order association names -- what a detach targets.
output "association_names" {
  description = "Names of the policy's associations, in declaration order"
  value       = local.association_names
}
