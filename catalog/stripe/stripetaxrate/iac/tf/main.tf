# StripeTaxRate Main Resources
#
# stripe_tax_rate is a manual tax rate. percentage and inclusive force a replacement, which
# deactivates the old rate and creates a new one with a new id; everything else (display_name,
# description, country, state, jurisdiction, tax_type, active and metadata) updates in place.
#
# Destroy deactivates the rate (active = false); Stripe keeps it, and it still applies to the
# subscriptions and invoices that already use it. The provider has no handling for a rate it
# cannot read: the next refresh fails, and the recovery is `tofu state rm` followed by an apply.
resource "stripe_tax_rate" "this" {
  display_name = var.spec.display_name
  percentage   = local.percentage
  inclusive    = local.inclusive
  country      = local.country
  state        = local.state
  jurisdiction = local.jurisdiction
  description  = local.description
  tax_type     = local.tax_type
  active       = local.active
  metadata     = local.metadata
}
