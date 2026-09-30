# StripePaymentLink Main Resources
#
# stripe_payment_link is a Stripe-hosted payment page at a public address. It takes payments from
# the moment it is created.
#
# Two line-item changes are accepted by the provider and never sent:
# - a price: Stripe's update call takes only a line item's id and quantity, so a changed price
#   would fail at apply or show as a pending change forever;
# - a quantity: the provider marks it write-only, so it is never stored, a change to it plans
#   nothing, and Stripe keeps the old quantity.
# The line_items tracker below holds each line item's price and quantity, in order, and a change
# to it replaces the link. With create_before_destroy the new link (and its new address, the url
# output other resources read by reference) exists before the old one is deactivated. Adjustable
# quantities and optional items stay in-place updates: the provider stores and sends them.
#
# The provider itself forces a replacement for currency, consent_collection, managed_payments,
# payment_intent_data.capture_method and setup_future_usage, shipping_options,
# subscription_data.description, and the Connect fields (application_fee_amount,
# application_fee_percent, on_behalf_of, transfer_data). Everything else updates in place.
# line_items[*].price_data (an inline price) is never set: the price it would create is invisible
# to this module and never archived -- the price is a StripePrice.
#
# Destroy deactivates the link (active = false); Stripe keeps it, and its address shows
# inactive_message. The provider has no handling for a link it cannot read: the next refresh
# fails, and the recovery is `tofu state rm` followed by an apply.

# The tracker holds nothing in Stripe and is never imported: an imported tracker has no input, so
# the first apply after an import would see the line items change and replace the imported link. Left
# out of the import, it is simply created on that first apply, and the link is untouched.
resource "terraform_data" "line_items" {
  input = local.line_items_tracked
}

resource "stripe_payment_link" "this" {
  line_items       = local.line_items
  optional_items   = local.optional_items
  active           = local.active
  inactive_message = local.inactive_message
  metadata         = local.metadata

  after_completion            = local.after_completion
  allow_promotion_codes       = try(var.spec.allow_promotion_codes, null)
  automatic_tax               = local.automatic_tax
  billing_address_collection  = local.billing_address_collection
  consent_collection          = local.consent_collection
  currency                    = local.currency
  custom_fields               = local.custom_fields
  custom_text                 = local.custom_text
  customer_creation           = local.customer_creation
  invoice_creation            = local.invoice_creation
  managed_payments            = local.managed_payments
  name_collection             = local.name_collection
  payment_intent_data         = local.payment_intent_data
  payment_method_collection   = local.payment_method_collection
  payment_method_options      = local.payment_method_options
  payment_method_types        = local.payment_method_types
  phone_number_collection     = local.phone_number_collection
  restrictions                = local.restrictions
  shipping_address_collection = local.shipping_address_collection
  shipping_options            = local.shipping_options
  submit_type                 = local.submit_type
  subscription_data           = local.subscription_data
  tax_id_collection           = local.tax_id_collection

  application_fee_amount  = try(var.spec.application_fee_amount, null)
  application_fee_percent = try(var.spec.application_fee_percent, null)
  on_behalf_of            = local.on_behalf_of
  transfer_data           = local.transfer_data

  lifecycle {
    create_before_destroy = true
    replace_triggered_by  = [terraform_data.line_items]
  }
}
