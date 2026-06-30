# Auth0EventStream Main Resources
# This file creates the Auth0 Event Stream

# Auth0 Event Stream Resource
resource "auth0_event_stream" "this" {
  name             = local.stream_name
  destination_type = local.destination_type
  subscriptions    = local.subscriptions

  # EventBridge configuration
  # Only applicable when destination_type is "eventbridge"
  # EventBridge configurations CANNOT be updated after creation
  dynamic "eventbridge_configuration" {
    for_each = local.destination_type == "eventbridge" && local.eventbridge_configuration != null ? [local.eventbridge_configuration] : []
    content {
      aws_account_id = eventbridge_configuration.value.aws_account_id
      aws_region     = eventbridge_configuration.value.aws_region
    }
  }

  # Webhook configuration
  # Only applicable when destination_type is "webhook"
  # Webhook configurations CAN be updated after creation
  dynamic "webhook_configuration" {
    for_each = local.destination_type == "webhook" && local.webhook_configuration != null ? [local.webhook_configuration] : []
    content {
      webhook_endpoint = webhook_configuration.value.webhook_endpoint

      webhook_authorization {
        method = webhook_configuration.value.webhook_authorization.method

        # Basic authentication credentials
        username = webhook_configuration.value.webhook_authorization.method == "basic" ? webhook_configuration.value.webhook_authorization.username : null
        password = webhook_configuration.value.webhook_authorization.method == "basic" ? webhook_configuration.value.webhook_authorization.password : null

        # Bearer token authentication
        token = webhook_configuration.value.webhook_authorization.method == "bearer" ? webhook_configuration.value.webhook_authorization.token : null
      }
    }
  }
}

