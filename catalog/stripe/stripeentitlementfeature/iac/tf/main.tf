# StripeEntitlementFeature Main Resources
#
# stripe_entitlements_feature is one capability a customer can be entitled to. lookup_key is sent
# only at create, so changing it replaces the feature; name and metadata update in place. The
# provider never sends active, so nothing here can reactivate an archived feature.
#
# Destroy archives the feature (active = false); Stripe keeps it. The provider has no handling for
# a feature it cannot read: the next refresh fails, and the recovery is `tofu state rm` followed
# by an apply, which creates a new feature.
resource "stripe_entitlements_feature" "this" {
  lookup_key = var.spec.lookup_key
  name       = var.spec.name
  metadata   = local.metadata
}
