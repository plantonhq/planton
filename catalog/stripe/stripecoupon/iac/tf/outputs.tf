# StripeCoupon Outputs
# Maps to the StripeCouponStackOutputs protobuf message.

output "id" {
  description = "The coupon's Stripe id; it changes when the coupon is replaced"
  value       = stripe_coupon.this.id
}

output "valid" {
  description = "Whether the coupon can still be applied to new customers"
  value       = stripe_coupon.this.valid
}
