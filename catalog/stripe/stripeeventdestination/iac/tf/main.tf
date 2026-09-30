# StripeEventDestination Main Resources
#
# stripe_v2_core_event_destination is where the account sends its events through Stripe's v2
# event destinations. type is derived from the destination block the spec sets, and with
# event_payload, events_from, snapshot_api_version and the amazon_eventbridge and
# azure_event_grid blocks it is sent only at create, so changing any of them replaces the
# destination. name, description, enabled_events, metadata and a webhook's url update in place.
#
# The provider always asks Stripe for a webhook destination's signing secret at create and keeps
# it in state from then on; an imported destination never has one. include is never set: it is
# write-only in the provider and changes nothing a later read returns. The API returns no AWS
# region, so the provider keeps the configured one -- an imported EventBridge destination has none
# in state, and its first apply replaces it.
#
# Destroy deletes the destination. The provider has no handling for a destination it cannot read:
# the next refresh fails, and the recovery is `tofu state rm` followed by an apply, which creates a
# new destination (a webhook destination with a new signing secret).
resource "stripe_v2_core_event_destination" "this" {
  name                 = var.spec.name
  description          = local.description
  type                 = local.type
  event_payload        = var.spec.event_payload
  enabled_events       = var.spec.enabled_events
  events_from          = local.events_from
  snapshot_api_version = local.snapshot_api_version
  metadata             = local.metadata
  azure_event_grid     = local.azure_event_grid

  dynamic "webhook_endpoint" {
    for_each = local.webhook_endpoint == null ? [] : [local.webhook_endpoint]
    content {
      url = webhook_endpoint.value.url
    }
  }

  dynamic "amazon_eventbridge" {
    for_each = local.amazon_eventbridge == null ? [] : [local.amazon_eventbridge]
    content {
      aws_account_id = amazon_eventbridge.value.aws_account_id
      aws_region     = amazon_eventbridge.value.aws_region
    }
  }
}
