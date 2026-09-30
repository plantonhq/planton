# Local values for the StripePrice module.
#
# An optional value the spec leaves at its zero value renders as null, so the provider never
# sends it and Stripe keeps its own default. Enums arrive as their value names, which are Stripe's
# own strings (per_unit, graduated, month, metered, inclusive). product arrives as a resolved
# product id: a reference to a StripeProduct is resolved before this module runs.
locals {
  unit_amount         = try(var.spec.unit_amount, null)
  unit_amount_decimal = try(var.spec.unit_amount_decimal, "") != "" ? var.spec.unit_amount_decimal : null
  billing_scheme      = try(var.spec.billing_scheme, "") != "" ? var.spec.billing_scheme : null
  tiers_mode          = try(var.spec.tiers_mode, "") != "" ? var.spec.tiers_mode : null
  lookup_key          = try(var.spec.lookup_key, "") != "" ? var.spec.lookup_key : null
  transfer_lookup_key = try(var.spec.transfer_lookup_key, false) ? true : null
  nickname            = try(var.spec.nickname, "") != "" ? var.spec.nickname : null
  tax_behavior        = try(var.spec.tax_behavior, "") != "" ? var.spec.tax_behavior : null
  metadata            = length(try(var.spec.metadata, {})) > 0 ? var.spec.metadata : null
  # The spec's default is active; the provider needs the value only to archive.
  active = try(var.spec.active, null) == null ? true : var.spec.active

  tiers = [
    for t in try(var.spec.tiers, []) : {
      up_to               = t.up_to
      flat_amount         = try(t.flat_amount, null)
      flat_amount_decimal = try(t.flat_amount_decimal, "") != "" ? t.flat_amount_decimal : null
      unit_amount         = try(t.unit_amount, null)
      unit_amount_decimal = try(t.unit_amount_decimal, "") != "" ? t.unit_amount_decimal : null
    }
  ]

  custom_unit_amount = try(var.spec.custom_unit_amount, null)

  transform_quantity = try(var.spec.transform_quantity, null) == null ? null : {
    divide_by = var.spec.transform_quantity.divide_by
    round     = var.spec.transform_quantity.round
  }

  recurring = try(var.spec.recurring, null) == null ? null : {
    interval          = var.spec.recurring.interval
    interval_count    = try(var.spec.recurring.interval_count, null)
    usage_type        = try(var.spec.recurring.usage_type, "") != "" ? var.spec.recurring.usage_type : null
    meter             = try(var.spec.recurring.meter, "") != "" ? var.spec.recurring.meter : null
    trial_period_days = try(var.spec.recurring.trial_period_days, null)
  }

  # Sorted by currency so the provider's list never reorders between plans.
  currency_options = [
    for currency in sort(keys(try(var.spec.currency_options, {}))) : {
      key                 = currency
      unit_amount         = try(var.spec.currency_options[currency].unit_amount, null)
      unit_amount_decimal = try(var.spec.currency_options[currency].unit_amount_decimal, "") != "" ? var.spec.currency_options[currency].unit_amount_decimal : null
      tax_behavior        = try(var.spec.currency_options[currency].tax_behavior, "") != "" ? var.spec.currency_options[currency].tax_behavior : null
      custom_unit_amount  = try(var.spec.currency_options[currency].custom_unit_amount, null)
      tiers = length(try(var.spec.currency_options[currency].tiers, [])) == 0 ? null : [
        for t in var.spec.currency_options[currency].tiers : {
          up_to               = t.up_to
          flat_amount         = try(t.flat_amount, null)
          flat_amount_decimal = try(t.flat_amount_decimal, "") != "" ? t.flat_amount_decimal : null
          unit_amount         = try(t.unit_amount, null)
          unit_amount_decimal = try(t.unit_amount_decimal, "") != "" ? t.unit_amount_decimal : null
        }
      ]
    }
  ]
}
