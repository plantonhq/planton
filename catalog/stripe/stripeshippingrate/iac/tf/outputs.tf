# StripeShippingRate Outputs
# Maps to the StripeShippingRateOutputs protobuf message.

output "id" {
  description = "The shipping rate's Stripe id (shr_...); it changes when the rate is replaced"
  value       = stripe_shipping_rate.this.id
}

output "active" {
  description = "false once the rate is deactivated"
  value       = stripe_shipping_rate.this.active
}
