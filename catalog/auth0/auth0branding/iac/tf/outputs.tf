# Auth0Branding Outputs
# Maps to the Auth0BrandingOutputs protobuf message: the branding as
# applied.

output "theme_id" {
  description = "The id of the theme the branding applied; empty when the spec declares no theme"
  value       = try(auth0_branding_theme.this[0].id, "")
}

output "logo_url" {
  description = "The logo the tenant's pages show after the deployment; empty when the spec manages no branding setting"
  value       = try(auth0_branding.this[0].logo_url, "")
}
