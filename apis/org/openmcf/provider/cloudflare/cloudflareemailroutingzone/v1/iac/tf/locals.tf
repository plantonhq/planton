locals {
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-email-routing-zone")

  catch_all = try(var.spec.catch_all, null)

  # Map the typed catch-all action onto the provider's generic {type, value[]}:
  # forward -> the destination addresses; worker -> the single script name;
  # drop -> no values.
  catch_all_values = local.catch_all == null ? [] : (
    local.catch_all.type == "forward" ? local.catch_all.forward_to : (
      local.catch_all.type == "worker" ? [local.catch_all.worker] : []
    )
  )
}
