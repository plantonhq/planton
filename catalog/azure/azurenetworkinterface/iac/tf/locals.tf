locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurenetworkinterface"
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

  # Metadata-derived tags first, then the user's spec tags merged over them:
  # user tags deliberately win so an org's governance conventions (cost
  # center, owner) can override the derived values where they collide.
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # Map the spec enums' name strings to ARM values. Unset applies Azure's
  # defaults (Dynamic allocation, IPv4), so an unspecified spec and Azure's
  # default deploy identically on both engines.
  ip_configurations = [
    for c in var.spec.ip_configurations : {
      name                 = c.name
      subnet_id            = c.subnet_id
      allocation           = c.private_ip_allocation == "STATIC" ? "Static" : "Dynamic"
      private_ip_address   = c.private_ip_address
      version              = c.private_ip_version == "IPV6" ? "IPv6" : "IPv4"
      public_ip_address_id = c.public_ip_address_id
      primary              = c.primary
      gateway_lb_fip_id    = c.gateway_load_balancer_frontend_ip_configuration_id
    }
  ]

  # Flatten each configuration's load-balancer / App Gateway memberships
  # into one collection per association type, keyed "{config}/{index}",
  # so a single for_each realizes every membership as its own ARM
  # association (joining/leaving never touches the NIC itself).
  lb_pool_associations = merge([
    for c in var.spec.ip_configurations : {
      for i, pool_id in c.load_balancer_backend_address_pool_ids : "${c.name}/${i}" => {
        ip_configuration_name = c.name
        pool_id               = pool_id
      }
    }
  ]...)

  lb_nat_rule_associations = merge([
    for c in var.spec.ip_configurations : {
      for i, rule_id in c.load_balancer_inbound_nat_rule_ids : "${c.name}/${i}" => {
        ip_configuration_name = c.name
        nat_rule_id           = rule_id
      }
    }
  ]...)

  appgw_pool_associations = merge([
    for c in var.spec.ip_configurations : {
      for i, pool_id in c.application_gateway_backend_address_pool_ids : "${c.name}/${i}" => {
        ip_configuration_name = c.name
        pool_id               = pool_id
      }
    }
  ]...)

  # ARM auxiliary values: AcceleratedConnections/Floating/MaxConnections
  # and A1/A2/A4/A8; null sends nothing (the non-appliance default).
  auxiliary_mode = (
    var.spec.auxiliary_mode == "ACCELERATED_CONNECTIONS" ? "AcceleratedConnections" :
    var.spec.auxiliary_mode == "FLOATING" ? "Floating" :
    var.spec.auxiliary_mode == "MAX_CONNECTIONS" ? "MaxConnections" : null
  )
  auxiliary_sku = (
    var.spec.auxiliary_sku == "A1" ? "A1" :
    var.spec.auxiliary_sku == "A2" ? "A2" :
    var.spec.auxiliary_sku == "A4" ? "A4" :
    var.spec.auxiliary_sku == "A8" ? "A8" : null
  )
}
