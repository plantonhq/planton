# StripeEntitlementFeature Outputs
# Maps to the StripeEntitlementFeatureOutputs protobuf message.

output "id" {
  description = "The feature's Stripe id (feat_...), the value a StripeProduct's features reference"
  value       = stripe_entitlements_feature.this.id
}

output "lookup_key" {
  description = "The feature's stable name, as the application checks it"
  value       = stripe_entitlements_feature.this.lookup_key
}

output "active" {
  description = "false once the feature is archived"
  value       = stripe_entitlements_feature.this.active
}
