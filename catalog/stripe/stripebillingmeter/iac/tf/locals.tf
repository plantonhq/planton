# Local values for the StripeBillingMeter module.
#
# An optional value the spec leaves at its zero value renders as null, so the provider never
# sends it and Stripe keeps its own default. Enums arrive as their value names, which are Stripe's
# own strings (sum, hour, one_time).
locals {
  event_time_window = try(var.spec.event_time_window, "") != "" ? var.spec.event_time_window : null
  customer_mapping  = try(var.spec.customer_mapping, null)
  value_settings    = try(var.spec.value_settings, null)

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
