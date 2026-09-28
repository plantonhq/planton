# Auth0TenantSettings Outputs
# Maps to the Auth0TenantSettingsStackOutputs protobuf message: the settings as
# the tenant carries them after the apply, managed or not. The Pulumi module's
# outputs.go exports the same names -- keep them in lockstep.

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

output "default_custom_domain" {
  description = "The tenant's default domain as set by this resource; empty when unmanaged"
  value       = try(auth0_custom_domain_default.this[0].domain, "")
}

output "default_audience" {
  description = "The API identifier every access token defaults to; empty when the tenant has none"
  value       = auth0_tenant.this.default_audience
}

output "default_directory" {
  description = "The connection the password grant signs people in through by default; empty when the tenant has none"
  value       = auth0_tenant.this.default_directory
}

output "client_id_metadata_document_supported" {
  description = "Whether the tenant registers applications from a Client ID Metadata Document"
  value       = auth0_tenant.this.client_id_metadata_document_supported
}

output "resource_parameter_profile" {
  description = "How a client names the API it wants a token for: audience or compatibility"
  value       = auth0_tenant.this.resource_parameter_profile
}

# flags is Optional+Computed: the provider always reads it back, so its one
# element exists after every apply.
output "enable_dynamic_client_registration" {
  description = "Whether any client can register a third-party application through /oidc/register"
  value       = try(auth0_tenant.this.flags[0].enable_dynamic_client_registration, false)
}

output "dynamic_client_registration_security_mode" {
  description = "The security mode of the applications Dynamic Client Registration creates"
  value       = auth0_tenant.this.dynamic_client_registration_security_mode
}

output "session_lifetime" {
  description = "The hours a login session lasts however active the person is"
  value       = auth0_tenant.this.session_lifetime
}

output "idle_session_lifetime" {
  description = "The hours a login session survives unused"
  value       = auth0_tenant.this.idle_session_lifetime
}

output "session_cookie_mode" {
  description = "Whether the session outlives the browser: persistent or non-persistent"
  value       = try(auth0_tenant.this.session_cookie[0].mode, "")
}

output "enabled_locales" {
  description = "The tenant's languages, its default first"
  value       = auth0_tenant.this.enabled_locales
}
