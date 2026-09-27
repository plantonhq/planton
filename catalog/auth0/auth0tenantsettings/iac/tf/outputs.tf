# Auth0TenantSettings Outputs
# Maps to the Auth0TenantSettingsStackOutputs protobuf message: the settings as
# the tenant carries them after the apply, managed or not.

output "friendly_name" {
  description = "The tenant's name as people see it"
  value       = auth0_tenant.this.friendly_name
}

output "picture_url" {
  description = "The URL of the tenant's logo"
  value       = auth0_tenant.this.picture_url
}

output "support_email" {
  description = "The support address the tenant's pages offer"
  value       = auth0_tenant.this.support_email
}

output "support_url" {
  description = "The support page the tenant's pages link to"
  value       = auth0_tenant.this.support_url
}
