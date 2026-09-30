# StripeProduct Outputs
# Maps to the StripeProductStackOutputs protobuf message.

output "id" {
  description = "The product's Stripe id (prod_...), the value a StripePrice's product references"
  value       = stripe_product.this.id
}

output "active" {
  description = "false once the product is archived"
  value       = stripe_product.this.active
}

output "default_price" {
  description = "The price Stripe treats as the product's default, when one is set (never set by this module)"
  value       = stripe_product.this.default_price
}

output "product_feature_ids" {
  description = "Each granted entitlement feature's id mapped to the id of the link that attaches it"
  value       = { for feature_id, link in stripe_product_feature.this : feature_id => link.id }
}
