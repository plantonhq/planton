# Local values for Auth0EventStream module
# These values are computed from the input variables

locals {
  # Core configuration
  stream_name      = var.metadata.name
  destination_type = var.spec.destination_type
  subscriptions    = var.spec.subscriptions

  # EventBridge configuration
  eventbridge_configuration = var.spec.eventbridge_configuration

  # Webhook configuration
  webhook_configuration = var.spec.webhook_configuration
}

