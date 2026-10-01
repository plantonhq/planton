# Local values for the StripeEntitlementFeature module.
#
# An empty metadata map renders as null, so the provider never sends it and Stripe keeps whatever
# keys the feature carries.
locals {
  metadata = length(try(var.spec.metadata, {})) > 0 ? var.spec.metadata : null
}
