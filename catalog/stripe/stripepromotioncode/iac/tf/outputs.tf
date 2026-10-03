# StripePromotionCode Outputs
# Maps to the StripePromotionCodeOutputs protobuf message.

output "id" {
  description = "The promotion code's Stripe id (promo_...); it changes when the code is replaced"
  value       = stripe_promotion_code.this.id
}

output "code" {
  description = "What customers type: the declared code, or the one Stripe generated"
  value       = stripe_promotion_code.this.code
}

output "active" {
  description = "false once the code is deactivated"
  value       = stripe_promotion_code.this.active
}
