# Local values for the StripeWebhookEndpoint module.
#
# An optional value the spec leaves at its zero value ("" or an empty map) renders as null, so
# the provider never sends it and Stripe keeps its own default. connect renders as null unless it
# is true: the provider stores connect exactly as configured and forces a replacement when it
# changes, so false and unset must mean the same thing.
locals {
  description = try(var.spec.description, "") != "" ? var.spec.description : null
  metadata    = length(try(var.spec.metadata, {})) > 0 ? var.spec.metadata : null
  api_version = try(var.spec.api_version, "") != "" ? var.spec.api_version : null
  connect     = try(var.spec.connect, false) ? true : null
}
