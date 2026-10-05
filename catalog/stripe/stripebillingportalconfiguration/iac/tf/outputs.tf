# StripeBillingPortalConfiguration Outputs
# Maps to the StripeBillingPortalConfigurationOutputs protobuf message: the configuration a
# portal session names.

output "id" {
  description = "The configuration's Stripe id (bpc_...), passed as configuration when a portal session is created"
  value       = stripe_billing_portal_configuration.this.id
}

output "is_default" {
  description = "Whether this is the account's default configuration"
  value       = stripe_billing_portal_configuration.this.is_default
}

output "active" {
  description = "Whether portal sessions may use the configuration"
  value       = stripe_billing_portal_configuration.this.active
}

output "login_page_url" {
  description = "The shareable portal sign-in URL, when login_page is enabled"
  value       = try(stripe_billing_portal_configuration.this.login_page.url, null)
}
