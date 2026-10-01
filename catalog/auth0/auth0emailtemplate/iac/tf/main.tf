# Auth0EmailTemplate Main Resources
#
# auth0_email_template customizes one of the emails the tenant the provider's
# credential belongs to sends; the template name is its identity. It sends
# through the tenant's email provider (Auth0EmailProvider): Auth0 refuses
# custom templates without one. When the template already exists, the create
# takes it over and rewrites it. Auth0 cannot delete a template: destroy
# disables it (a PATCH of enabled = false), and the tenant sends Auth0's
# default email again.
resource "auth0_email_template" "this" {
  template                  = var.spec.template
  from                      = var.spec.from
  subject                   = var.spec.subject
  body                      = var.spec.body
  syntax                    = local.syntax
  result_url                = local.result_url
  url_lifetime_in_seconds   = local.url_lifetime_in_seconds
  enabled                   = local.enabled
  include_email_in_redirect = local.include_email_in_redirect
}
