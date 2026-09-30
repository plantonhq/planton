# StripeWebhookEndpoint Main Resources
#
# stripe_webhook_endpoint is where the account the provider's key belongs to delivers its
# events. Stripe returns the endpoint's signing secret only in the create response; the provider
# keeps it in state from then on, and never recovers it on import. api_version and connect are
# sent only at create, so changing either replaces the endpoint -- and a new endpoint means a new
# signing secret, which the receiving service must take in the same deployment. url,
# enabled_events, description and metadata update in place.
#
# Destroy deletes the endpoint. The provider has no handling for an endpoint deleted outside
# Planton: the next refresh fails reading it, and the recovery is `tofu state rm` followed by an
# apply, which creates a new endpoint with a new secret.
resource "stripe_webhook_endpoint" "this" {
  url            = var.spec.url
  enabled_events = var.spec.enabled_events
  description    = local.description
  metadata       = local.metadata
  api_version    = local.api_version
  connect        = local.connect
}
