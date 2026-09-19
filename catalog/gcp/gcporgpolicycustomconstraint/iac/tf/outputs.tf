# The constraint's full resource name,
# organizations/{org}/customConstraints/custom.{name}. The provider's `name`
# attribute is the custom.{name} handle; the resource id is the full name.
output "name" {
  description = "The custom constraint's full resource name"
  value       = google_org_policy_custom_constraint.this.id
}

# The custom.{name} handle a GcpOrgPolicy enforces.
output "constraint" {
  description = "The constraint handle (custom.{name}) a policy references"
  value       = google_org_policy_custom_constraint.this.name
}

output "update_time" {
  description = "When the constraint was last updated (RFC 3339)"
  value       = google_org_policy_custom_constraint.this.update_time
}
