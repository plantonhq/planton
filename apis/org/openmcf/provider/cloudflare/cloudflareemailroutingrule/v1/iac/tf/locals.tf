locals {
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-email-routing-rule")

  # Map the typed action onto the provider's generic {type, value[]}.
  action_values = var.spec.action.type == "forward" ? var.spec.action.forward_to : (
    var.spec.action.type == "worker" ? [var.spec.action.worker] : []
  )
}
