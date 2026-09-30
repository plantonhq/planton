# StripeTaxRegistration Main Resources
#
# stripe_tax_registration tells Stripe Tax the account is registered to collect tax in one place.
# From active_from, Stripe Tax calculates and collects tax there on every payment that uses
# automatic tax.
#
# Only active_from and expires_at update in place: they are the only arguments the provider sends
# on an update. country and country_options force a replacement, which creates a new registration
# and forgets the old one while it is still active in Stripe. expires_at keeps its stored value
# when the manifest drops it (the provider reuses state for an unknown value), so a new date is
# the only change an expiry takes.
#
# Destroy only removes the registration from state: the provider makes no call, and Stripe keeps
# collecting until expires_at passes. The provider has no handling for a registration it cannot
# read: the next refresh fails, and the recovery is `tofu state rm` followed by an apply.
resource "stripe_tax_registration" "this" {
  country         = var.spec.country
  active_from     = var.spec.active_from
  expires_at      = local.expires_at
  country_options = { (local.country_key) = local.country_option }
}
