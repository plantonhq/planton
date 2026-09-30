# StripeBillingMeter Main Resources
#
# stripe_billing_meter counts usage events. Only display_name updates in place; event_name,
# event_time_window, customer_mapping, default_aggregation and value_settings force a replacement,
# which deactivates the old meter and creates a new one with a new id. A metered StripePrice that
# references the meter is replaced on the same apply, and an application that reads event_name by
# reference follows it.
#
# customer_mapping.type is always by_id: it is the only mapping Stripe offers, so the module sets
# it and the spec names only the payload key.
#
# Destroy deactivates the meter (status inactive); Stripe keeps it. The provider has no handling
# for a meter it cannot read: the next refresh fails, and the recovery is `tofu state rm` followed
# by an apply.
resource "stripe_billing_meter" "this" {
  display_name      = var.spec.display_name
  event_name        = var.spec.event_name
  event_time_window = local.event_time_window

  default_aggregation {
    formula = var.spec.default_aggregation.formula
  }

  dynamic "customer_mapping" {
    for_each = local.customer_mapping == null ? [] : [local.customer_mapping]
    content {
      type              = "by_id"
      event_payload_key = customer_mapping.value.event_payload_key
    }
  }

  dynamic "value_settings" {
    for_each = local.value_settings == null ? [] : [local.value_settings]
    content {
      event_payload_key = value_settings.value.event_payload_key
    }
  }
}

# stripe_billing_alert is a usage threshold on the meter. Every field forces a replacement, and
# the provider's delete only forgets the alert: Stripe offers no delete, so a replaced, removed or
# destroyed alert stays active in Stripe until it is archived in the Dashboard or through the API.
# alert_type is always usage_threshold, the only alert type Stripe offers, and a customer filter's
# type is always customer; the module sets both.
resource "stripe_billing_alert" "this" {
  for_each = local.alerts

  alert_type = "usage_threshold"
  title      = each.key

  usage_threshold = {
    meter      = stripe_billing_meter.this.id
    gte        = each.value.gte
    recurrence = each.value.recurrence
    filters    = each.value.customer == null ? null : [{ type = "customer", customer = each.value.customer }]
  }
}
