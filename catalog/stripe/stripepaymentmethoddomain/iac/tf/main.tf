# StripePaymentMethodDomain Main Resources
#
# stripe_payment_method_domain registers a domain for wallet buttons. domain_name is sent only at
# create, so changing it replaces the registration; enabled updates in place. The provider never
# calls Stripe's validate endpoint: Stripe validates the domain for each wallet on its own, and the
# per-wallet status outputs report the result.
#
# Destroy only removes the registration from state -- the provider makes no API call, and the
# domain stays registered, and enabled, in Stripe. The provider has no handling for a domain it
# cannot read: the next refresh fails, and the recovery is `tofu state rm` followed by an apply.
resource "stripe_payment_method_domain" "this" {
  domain_name = var.spec.domain_name
  enabled     = local.enabled
}
