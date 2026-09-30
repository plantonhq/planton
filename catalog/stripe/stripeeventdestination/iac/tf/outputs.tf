# StripeEventDestination Outputs
# Maps to the StripeEventDestinationStackOutputs protobuf message: the destination's id, a
# webhook destination's signing secret, and the cloud-side source an EventBridge or Event Grid
# destination waits on.

output "id" {
  description = "The destination's Stripe id"
  value       = stripe_v2_core_event_destination.this.id
}

output "status" {
  description = "enabled or disabled, as Stripe reports it"
  value       = stripe_v2_core_event_destination.this.status
}

output "status_disabled_reason" {
  description = "Why a disabled destination is disabled"
  value       = try(stripe_v2_core_event_destination.this.status_details.disabled.reason, null)
}

output "signing_secret" {
  description = "A webhook destination's signing secret -- empty for an imported destination, since Stripe returns it only at creation"
  # Reads the secret leaf directly (never a conditional over the whole resource) so the output
  # carries only the leaf's own sensitivity.
  value     = try(stripe_v2_core_event_destination.this.webhook_endpoint[0].signing_secret, null)
  sensitive = true
}

output "aws_event_source_arn" {
  description = "The ARN of the partner event source an EventBridge destination created"
  value       = local.aws_event_source_arn
}

output "aws_event_source_name" {
  description = "The partner event source's name (aws.partner/stripe.com/...), the value an EventBridge event bus associates with"
  value       = local.aws_event_source_arn == null ? null : try(regex("event-source/(.+)$", local.aws_event_source_arn)[0], null)
}

output "aws_event_source_status" {
  description = "pending until the source is associated with an event bus, then active"
  value       = try(stripe_v2_core_event_destination.this.amazon_eventbridge[0].aws_event_source_status, null)
}

output "azure_partner_topic_name" {
  description = "The partner topic an Event Grid destination created"
  value       = try(stripe_v2_core_event_destination.this.azure_event_grid.azure_partner_topic_name, null)
}

output "azure_partner_topic_status" {
  description = "never_activated until the partner topic is activated in Azure, then activated"
  value       = try(stripe_v2_core_event_destination.this.azure_event_grid.azure_partner_topic_status, null)
}

locals {
  aws_event_source_arn = try(stripe_v2_core_event_destination.this.amazon_eventbridge[0].aws_event_source_arn, null)
}
