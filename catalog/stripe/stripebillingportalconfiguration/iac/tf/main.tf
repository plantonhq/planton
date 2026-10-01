# StripeBillingPortalConfiguration Main Resources
#
# stripe_billing_portal_configuration is one customer-portal configuration in the account the
# provider's key belongs to. It is always a new configuration: the module never adopts the
# account's default, and the application names this one's id when it opens a portal session.
#
# Stripe's create call cannot set active, so the provider makes a second, update call right after
# create when active is false. Stripe never deletes a configuration: destroy is an update to
# active = false, skipped when the configuration is already inactive, and Stripe keeps it forever.
# The provider has no handling for a configuration that cannot be read; the next refresh fails
# until it is removed from state.
resource "stripe_billing_portal_configuration" "this" {
  name               = local.name
  default_return_url = local.default_return_url
  metadata           = local.metadata
  active             = local.active
  business_profile   = local.business_profile
  login_page         = local.login_page

  features = {
    customer_update       = local.customer_update
    invoice_history       = local.invoice_history
    payment_method_update = local.payment_method_update
    subscription_cancel   = local.subscription_cancel
    subscription_update   = local.subscription_update
  }
}
