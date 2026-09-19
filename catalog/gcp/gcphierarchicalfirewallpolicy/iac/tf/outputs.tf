# The policy's server-assigned numeric ID -- Google's `name` for a
# hierarchical policy, and what the firewall logs identify it by.
output "policy_id" {
  description = "The policy's server-assigned numeric ID"
  value       = google_compute_firewall_policy.this.name
}

# The user-facing short name, as declared or defaulted from metadata.name.
output "short_name" {
  description = "The policy's short name"
  value       = google_compute_firewall_policy.this.short_name
}

output "self_link" {
  description = "Self-link URL of the policy"
  value       = google_compute_firewall_policy.this.self_link
}

# organizations/{id} or folders/{id}.
output "parent" {
  description = "The node the policy lives under"
  value       = google_compute_firewall_policy.this.parent
}

# Google's complexity measure for the whole rule set; an organization has a
# quota of tuples across its hierarchical policies.
output "rule_tuple_count" {
  description = "Total rule tuple count of the policy"
  value       = google_compute_firewall_policy.this.rule_tuple_count
}

# Declaration-order association names -- what a detach targets.
output "association_names" {
  description = "Names of the policy's associations, in declaration order"
  value       = local.association_names
}
