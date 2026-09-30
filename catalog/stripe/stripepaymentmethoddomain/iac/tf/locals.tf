# Local values for the StripePaymentMethodDomain module.
locals {
  # The spec's default is enabled; the value is always sent so a registration disabled in the
  # Dashboard is enabled again by the next apply unless the spec says otherwise.
  enabled = try(var.spec.enabled, null) == null ? true : var.spec.enabled
}
