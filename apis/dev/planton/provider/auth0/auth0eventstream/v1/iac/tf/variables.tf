# Auth0EventStream Variables
# This file defines all input variables for the Auth0EventStream Terraform module
# These variables map to the Auth0EventStreamSpec protobuf message

variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Auth0EventStream specification"
  type = object({
    # destination_type specifies where events should be delivered.
    # Required. One of: eventbridge, webhook
    destination_type = string

    # subscriptions is a list of event types this stream is subscribed to.
    # Required. At least one subscription is required.
    # Common event types: user.created, user.updated, authentication.success, authentication.failure
    subscriptions = list(string)

    # eventbridge_configuration contains settings for AWS EventBridge destination.
    # Only applicable when destination_type is "eventbridge".
    # EventBridge configurations CANNOT be updated after creation.
    eventbridge_configuration = optional(object({
      # aws_account_id is the 12-digit AWS account ID where events will be delivered.
      aws_account_id = string

      # aws_region is the AWS region for the EventBridge event bus.
      aws_region = string
    }))

    # webhook_configuration contains settings for webhook destination.
    # Only applicable when destination_type is "webhook".
    # Webhook configurations CAN be updated after creation.
    webhook_configuration = optional(object({
      # webhook_endpoint is the HTTPS URL that will receive webhook events.
      # Must be a valid, publicly accessible HTTPS URL.
      webhook_endpoint = string

      # webhook_authorization contains authentication settings for the webhook endpoint.
      webhook_authorization = object({
        # method specifies the authorization method: basic or bearer
        method = string

        # username is for basic authentication
        username = optional(string)

        # password is for basic authentication
        password = optional(string)

        # token is for bearer token authentication
        token = optional(string)
      })
    }))
  })
}

