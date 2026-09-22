output "name" {
  description = "The group's resource name (groups/{group_id})"
  value       = google_cloud_identity_group.this.name
}

# The group's identity in IAM bindings (group:{email}).
output "group_email" {
  description = "The group's email address"
  value       = var.spec.group_email
}

output "membership_count" {
  description = "Number of memberships this manifest manages"
  value       = tostring(length(var.spec.memberships))
}
