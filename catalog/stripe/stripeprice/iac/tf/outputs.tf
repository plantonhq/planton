# StripePrice Outputs
# Maps to the StripePriceStackOutputs protobuf message.

output "id" {
  description = "The price's Stripe id (price_...); it changes when the price is replaced"
  value       = stripe_price.this.id
}

output "type" {
  description = "one_time or recurring"
  value       = stripe_price.this.type
}

output "active" {
  description = "false once the price is archived"
  value       = stripe_price.this.active
}

output "lookup_key" {
  description = "The stable name the application retrieves the price by, when one is set"
  value       = stripe_price.this.lookup_key
}

output "product" {
  description = "The id of the product the price charges for"
  value       = stripe_price.this.product
}
