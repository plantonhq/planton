# Local values for the Auth0EmailTemplate module.
#
# The provider requires syntax and enabled, so the module fills the spec's
# defaults when they are unset: "liquid", the only language Auth0 renders, and
# on. Every other optional setting left unset renders as null, so the provider
# never sends it and Auth0 applies its own default; empty strings are the
# proto's zero value for "unset", so result_url maps "" to null here. The
# Pulumi module's locals.go applies the same rule -- keep them in lockstep.
locals {
  syntax  = coalesce(try(var.spec.syntax, null), "liquid")
  enabled = try(var.spec.enabled, null) == null ? true : var.spec.enabled

  result_url                = try(var.spec.result_url, "") != "" ? var.spec.result_url : null
  url_lifetime_in_seconds   = try(var.spec.url_lifetime_in_seconds, null)
  include_email_in_redirect = try(var.spec.include_email_in_redirect, null)
}
