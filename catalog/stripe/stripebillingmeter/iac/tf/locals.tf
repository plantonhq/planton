# Local values for the StripeBillingMeter module.
#
# An optional value the spec leaves at its zero value renders as null, so the provider never
# sends it and Stripe keeps its own default. Enums arrive as their value names, which are Stripe's
# own strings (sum, hour, one_time).
locals {
  event_time_window = try(var.spec.event_time_window, "") != "" ? var.spec.event_time_window : null
  # Stripe fills both payload keys with its defaults when a meter is created without them, and the
  # provider then rejects its own result (a block it did not plan appears; verified live). The
  # module therefore always sends both, with Stripe's own defaults when the spec leaves them unset,
  # so what is planned is what Stripe stores. A count meter ignores value_settings.
  customer_payload_key = try(var.spec.customer_mapping.event_payload_key, "") != "" ? var.spec.customer_mapping.event_payload_key : "stripe_customer_id"
  value_payload_key    = try(var.spec.value_settings.event_payload_key, "") != "" ? var.spec.value_settings.event_payload_key : "value"

  # Keyed by title (unique per meter, enforced by the spec), so adding or removing one alert
  # touches only that alert, and the alert_ids output can address each alert by its title.
  alerts = {
    for a in try(var.spec.alerts, []) : a.title => {
      gte        = a.gte
      customer   = try(a.customer, "") != "" ? a.customer : null
      recurrence = try(a.recurrence, "") != "" ? a.recurrence : "one_time"
    }
  }
}
