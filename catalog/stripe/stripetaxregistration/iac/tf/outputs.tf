# StripeTaxRegistration Outputs
# Maps to the StripeTaxRegistrationOutputs protobuf message.

output "id" {
  description = "The registration's Stripe id (taxreg_...); it changes when the registration is replaced"
  value       = stripe_tax_registration.this.id
}

output "status" {
  description = "scheduled, active or expired, read from the registration's dates"
  value       = stripe_tax_registration.this.status
}
