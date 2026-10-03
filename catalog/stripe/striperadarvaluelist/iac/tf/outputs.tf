# StripeRadarValueList Outputs
# Maps to the StripeRadarValueListOutputs protobuf message.

output "id" {
  description = "The list's Stripe id (rsl_...)"
  value       = stripe_radar_value_list.this.id
}

output "alias" {
  description = "The name Radar rules reference the list by"
  value       = stripe_radar_value_list.this.alias
}

output "item_ids" {
  description = "Each item's value mapped to its Stripe id (rsli_...)"
  value       = { for value, item in stripe_radar_value_list_item.this : value => item.id }
}
