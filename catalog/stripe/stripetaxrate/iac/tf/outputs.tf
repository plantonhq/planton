# StripeTaxRate Outputs
# Maps to the StripeTaxRateOutputs protobuf message.

output "id" {
  description = "The tax rate's Stripe id (txr_...); it changes when the rate is replaced"
  value       = stripe_tax_rate.this.id
}

output "active" {
  description = "false once the rate is deactivated"
  value       = stripe_tax_rate.this.active
}
