# Create the VPN site -- the Virtual WAN address-book entry for one
# branch location. The object is free and provisions in seconds; it
# deploys nothing at the branch. Deleting a site requires the
# connections pointing at it to be gone first (the runner's reverse
# teardown handles the ordering).
resource "azurerm_vpn_site" "main" {
  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group
  virtual_wan_id      = var.spec.virtual_wan_id

  # The prefixes Azure routes into the tunnels. Empty is legal when
  # every link speaks BGP (the spec documents the convention).
  address_cidrs = var.spec.address_cidrs

  # The provider validates these as non-empty when configured -- emit
  # null instead of "" so an unset spec field stays unset on the wire.
  device_vendor = var.spec.device_vendor != "" ? var.spec.device_vendor : null
  device_model  = var.spec.device_model != "" ? var.spec.device_model : null

  # The branch's internet links -- ARM returns each link's ID, which the
  # link_ids output republishes keyed by name for connections to pin to.
  dynamic "link" {
    for_each = var.spec.links
    content {
      name          = link.value.name
      provider_name = link.value.provider_name != "" ? link.value.provider_name : null
      speed_in_mbps = link.value.speed_in_mbps

      # The spec guarantees at least one endpoint per link; null (not
      # "") keeps the provider's IsIPAddress/non-empty validations off
      # the unset one.
      ip_address = link.value.ip_address != "" ? link.value.ip_address : null
      fqdn       = link.value.fqdn != "" ? link.value.fqdn : null

      dynamic "bgp" {
        for_each = link.value.bgp != null ? [link.value.bgp] : []
        content {
          asn             = bgp.value.asn
          peering_address = bgp.value.peering_address
        }
      }
    }
  }

  # O365 breakout categories for SD-WAN partners. Unset sends nothing
  # (ARM's no-breakout default).
  dynamic "o365_policy" {
    for_each = var.spec.o365_policy != null ? [var.spec.o365_policy] : []
    content {
      dynamic "traffic_category" {
        for_each = o365_policy.value.traffic_category != null ? [o365_policy.value.traffic_category] : []
        content {
          allow_endpoint_enabled    = coalesce(traffic_category.value.allow_endpoint_enabled, false)
          default_endpoint_enabled  = coalesce(traffic_category.value.default_endpoint_enabled, false)
          optimize_endpoint_enabled = coalesce(traffic_category.value.optimize_endpoint_enabled, false)
        }
      }
    }
  }

  tags = local.final_tags
}
