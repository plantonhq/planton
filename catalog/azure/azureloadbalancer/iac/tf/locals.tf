locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azureloadbalancer"
    "resource_name" = var.metadata.name
  }

  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  id_tag = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "resource_id" = var.metadata.id } : {}

  # Metadata-derived tags first, then the user's spec tags merged over
  # them: user tags deliberately win so an org's governance conventions
  # (cost center, owner) can override the derived values where they
  # collide.
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # Map the spec enums' name strings to ARM values. Enum values arrive as
  # the FULL proto value name; unset applies Azure's defaults (Standard
  # SKU, Regional tier) so an unspecified spec and Azure's default deploy
  # identically on both engines.
  sku = (
    var.spec.sku == "GATEWAY" ? "Gateway" : "Standard"
  )
  sku_tier = (
    var.spec.sku_tier == "GLOBAL" ? "Global" : "Regional"
  )

  # Frontends with enum names mapped to ARM values. When the private
  # address is pinned the allocation flips to Static; internal frontends
  # without a pin stay Dynamic. Public frontends carry no allocation.
  frontend_ip_configurations = [
    for f in var.spec.frontend_ip_configurations : {
      name                 = f.name
      subnet_id            = f.subnet_id
      public_ip_address_id = f.public_ip_address_id
      public_ip_prefix_id  = f.public_ip_prefix_id
      private_ip_address   = f.private_ip_address
      is_internal          = f.subnet_id != null && f.subnet_id != ""
      allocation = (
        f.subnet_id != null && f.subnet_id != ""
        ? (f.private_ip_address != null && f.private_ip_address != "" ? "Static" : "Dynamic")
        : null
      )
      version           = f.private_ip_address_version == "IPV6" ? "IPv6" : (f.subnet_id != null && f.subnet_id != "" ? "IPv4" : null)
      zones             = f.zones
      gateway_lb_fip_id = f.gateway_load_balancer_frontend_ip_configuration_id
    }
  ]

  # The default frontend name: rules may omit the frontend when exactly
  # one is declared (spec-level validation guarantees the omission only
  # happens then).
  default_frontend_name = var.spec.frontend_ip_configurations[0].name

  # Transport protocols share one enum across rules, NAT rules, and
  # outbound rules; probe protocol and load distribution have their own.
  transport_protocol_map = {
    "TCP" = "Tcp"
    "UDP" = "Udp"
    "ALL" = "All"
  }
  probe_protocol_map = {
    "PROBE_TCP"   = "Tcp"
    "PROBE_HTTP"  = "Http"
    "PROBE_HTTPS" = "Https"
  }
  load_distribution_map = {
    "DEFAULT"            = "Default"
    "SOURCE_IP"          = "SourceIP"
    "SOURCE_IP_PROTOCOL" = "SourceIPProtocol"
  }
  tunnel_protocol_map = {
    "TUNNEL_PROTOCOL_NONE" = "None"
    "NATIVE"               = "Native"
    "VXLAN"                = "VXLAN"
  }
  tunnel_type_map = {
    "TUNNEL_TYPE_NONE" = "None"
    "INTERNAL"         = "Internal"
    "EXTERNAL"         = "External"
  }
  synchronous_mode_map = {
    "AUTOMATIC" = "Automatic"
    "MANUAL"    = "Manual"
  }

  # Name-keyed sub-resource lookups for cross-references. Rules resolve
  # pools/probes through these; a stale name fails the plan loudly, which
  # is the desired behavior (spec-level validation catches it first).
  backend_pool_map = { for pool in azurerm_lb_backend_address_pool.pools : pool.name => pool.id }
  probe_map        = { for probe in azurerm_lb_probe.probes : probe.name => probe.id }

  # Flatten pool addresses into one collection keyed "pool/address" so a
  # single for_each realizes every IP-based member.
  pool_addresses = merge([
    for pool in var.spec.backend_pools : {
      for addr in pool.addresses : "${pool.name}/${addr.name}" => {
        pool_name  = pool.name
        name       = addr.name
        ip_address = addr.ip_address
        lb_fip_id  = addr.load_balancer_frontend_ip_configuration_id
        vnet_id    = pool.virtual_network_id
      }
    }
  ]...)
}
