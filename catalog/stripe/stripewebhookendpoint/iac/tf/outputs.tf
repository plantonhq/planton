# StripeWebhookEndpoint Outputs
# Maps to the StripeWebhookEndpointStackOutputs protobuf message: the endpoint's id and signing
# secret.

output "id" {
  description = "The endpoint's Stripe id (we_...)"
  value       = stripe_webhook_endpoint.this.id
}

output "secret" {
  description = "The signing secret (whsec_...) that verifies each delivery -- empty for an imported endpoint, since Stripe returns it only at creation"
  # Reads the secret leaf directly (never a conditional over the whole resource) so the output
  # carries only the leaf's own sensitivity.
  value     = stripe_webhook_endpoint.this.secret
  sensitive = true
}

output "status" {
  description = "enabled or disabled, as Stripe reports it"
  value       = stripe_webhook_endpoint.this.status
}

output "url" {
  description = "The address events are delivered to"
  value       = stripe_webhook_endpoint.this.url
}

output "application" {
  description = "The Connect application that created the endpoint, when one did"
  value       = stripe_webhook_endpoint.this.application
}
