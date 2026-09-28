# Auth0TenantSettings Outputs
# Maps to the Auth0TenantSettingsStackOutputs protobuf message: the settings as
# the tenant carries them after the apply, managed or not (local.tenant is the
# resource when the spec declares a setting, the data source otherwise). The Pulumi module's
# outputs.go exports the same names -- keep them in lockstep.

output "friendly_name" {
  description = "The tenant's name as people see it"
  value       = local.tenant.friendly_name
}

output "picture_url" {
  description = "The URL of the tenant's logo"
  value       = local.tenant.picture_url
}

output "support_email" {
  description = "The support address the tenant's pages offer"
  value       = local.tenant.support_email
}

output "support_url" {
  description = "The support page the tenant's pages link to"
  value       = local.tenant.support_url
}

output "default_custom_domain" {
  description = "The tenant's default domain as set by this resource; empty when unmanaged"
  value       = try(auth0_custom_domain_default.this[0].domain, "")
}

output "default_audience" {
  description = "The API identifier every access token defaults to; empty when the tenant has none"
  value       = local.tenant.default_audience
}

output "default_directory" {
  description = "The connection the password grant signs people in through by default; empty when the tenant has none"
  value       = local.tenant.default_directory
}

output "client_id_metadata_document_supported" {
  description = "Whether the tenant registers applications from a Client ID Metadata Document"
  value       = local.tenant.client_id_metadata_document_supported
}

output "resource_parameter_profile" {
  description = "How a client names the API it wants a token for: audience or compatibility"
  value       = local.tenant.resource_parameter_profile
}

# flags is Optional+Computed: the provider always reads it back, so its one
# element exists after every apply.
output "enable_dynamic_client_registration" {
  description = "Whether any client can register a third-party application through /oidc/register"
  value       = try(local.tenant.flags[0].enable_dynamic_client_registration, false)
}

output "dynamic_client_registration_security_mode" {
  description = "The security mode of the applications Dynamic Client Registration creates"
  value       = local.tenant.dynamic_client_registration_security_mode
}

output "session_lifetime" {
  description = "The hours a login session lasts however active the person is"
  value       = local.tenant.session_lifetime
}

output "idle_session_lifetime" {
  description = "The hours a login session survives unused"
  value       = local.tenant.idle_session_lifetime
}

output "session_cookie_mode" {
  description = "Whether the session outlives the browser: persistent or non-persistent"
  value       = try(local.tenant.session_cookie[0].mode, "")
}

output "enabled_locales" {
  description = "The tenant's languages, its default first"
  value       = local.tenant.enabled_locales
}
