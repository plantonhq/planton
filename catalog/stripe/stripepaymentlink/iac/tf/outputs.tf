# StripePaymentLink Outputs
# Maps to the StripePaymentLinkOutputs protobuf message.

output "id" {
  description = "The payment link's Stripe id (plink_...); it changes when the link is replaced"
  value       = stripe_payment_link.this.id
}

output "url" {
  description = "The page's public address; it changes when the link is replaced"
  value       = stripe_payment_link.this.url
}

output "active" {
  description = "false once the link is deactivated"
  value       = stripe_payment_link.this.active
}
