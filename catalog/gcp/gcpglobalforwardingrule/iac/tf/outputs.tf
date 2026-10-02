# Exactly one of the two resources exists (see the count guards in main.tf),
# so each output selects whichever branch was created with one(concat(...)).

# The VIP — the IP address DNS records point at. The API always reports the
# literal IP number here, even when the spec referenced an address resource.
output "ip_address" {
  description = "IP address the forwarding rule accepts traffic on (the load balancer VIP)"
  value       = one(concat(google_compute_global_forwarding_rule.this[*].ip_address, google_compute_forwarding_rule.this[*].ip_address))
}

# Self-link URI of the forwarding rule. A regional rule's link carries
# regions/{region} where the global one says global.
output "self_link" {
  description = "Self-link URI of the forwarding rule"
  value       = one(concat(google_compute_global_forwarding_rule.this[*].self_link, google_compute_forwarding_rule.this[*].self_link))
}

# The name as it exists in GCP.
output "forwarding_rule_name" {
  description = "Name of the forwarding rule in GCP"
  value       = one(concat(google_compute_global_forwarding_rule.this[*].name, google_compute_forwarding_rule.this[*].name))
}

# Server-assigned numeric ID of the forwarding rule.
output "forwarding_rule_id" {
  description = "Server-assigned numeric ID of the forwarding rule"
  value       = one(concat(google_compute_global_forwarding_rule.this[*].forwarding_rule_id, google_compute_forwarding_rule.this[*].forwarding_rule_id))
}

# PSC connection id — populated only for Private Service Connect frontends.
output "psc_connection_id" {
  description = "Private Service Connect connection id (PSC frontends only)"
  value       = one(concat(google_compute_global_forwarding_rule.this[*].psc_connection_id, google_compute_forwarding_rule.this[*].psc_connection_id))
}

# PSC connection status — ACCEPTED means the producer admitted the
# connection. Populated only for Private Service Connect frontends.
output "psc_connection_status" {
  description = "Private Service Connect connection status (PSC frontends only)"
  value       = one(concat(google_compute_global_forwarding_rule.this[*].psc_connection_status, google_compute_forwarding_rule.this[*].psc_connection_status))
}

# Region of a regional rule; empty for a global one, so downstream
# composition can confirm scope compatibility.
output "region" {
  description = "Region of the forwarding rule (empty for global)"
  value       = local.is_regional ? var.spec.region : ""
}

# The internal DNS name of an internal passthrough Network Load Balancer
# that set service_label; empty otherwise (and always empty for a global
# rule, whose API collection has no such attribute).
output "service_name" {
  description = "Internal DNS service name of an internal passthrough NLB with service_label (empty otherwise)"
  value       = local.is_regional ? one(google_compute_forwarding_rule.this[*].service_name) : ""
}
