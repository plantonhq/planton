# Local values for the StripeEventDestination module.
#
# An optional value the spec leaves at its zero value renders as null, so the provider never
# sends it and Stripe keeps its own default. event_payload arrives as its enum value name, which
# is Stripe's own string (snapshot, thin). The spec sets exactly one destination block; Stripe's
# type is the name of that block.
locals {
  webhook_endpoint   = try(var.spec.webhook_endpoint, null)
  amazon_eventbridge = try(var.spec.amazon_eventbridge, null)
  azure_event_grid = try(var.spec.azure_event_grid, null) == null ? null : {
    azure_subscription_id     = var.spec.azure_event_grid.azure_subscription_id
    azure_resource_group_name = var.spec.azure_event_grid.azure_resource_group_name
    azure_region              = var.spec.azure_event_grid.azure_region
  }

  type = (
    local.webhook_endpoint != null ? "webhook_endpoint" :
    local.amazon_eventbridge != null ? "amazon_eventbridge" :
    "azure_event_grid"
  )

  description          = try(var.spec.description, "") != "" ? var.spec.description : null
  events_from          = length(try(var.spec.events_from, [])) > 0 ? var.spec.events_from : null
  snapshot_api_version = try(var.spec.snapshot_api_version, "") != "" ? var.spec.snapshot_api_version : null
  metadata             = length(try(var.spec.metadata, {})) > 0 ? var.spec.metadata : null
}
