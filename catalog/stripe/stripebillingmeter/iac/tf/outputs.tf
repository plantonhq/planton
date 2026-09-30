# StripeBillingMeter Outputs
# Maps to the StripeBillingMeterStackOutputs protobuf message.

output "id" {
  description = "The meter's Stripe id (mtr_...); it changes when the meter is replaced"
  value       = stripe_billing_meter.this.id
}

output "event_name" {
  description = "The name the application sends with each usage event"
  value       = stripe_billing_meter.this.event_name
}

output "status" {
  description = "active, or inactive once the meter is deactivated"
  value       = stripe_billing_meter.this.status
}

output "alert_ids" {
  description = "Each alert's title mapped to its Stripe id"
  value       = { for title, alert in stripe_billing_alert.this : title => alert.id }
}
