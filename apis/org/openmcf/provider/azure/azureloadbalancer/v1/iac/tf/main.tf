# Create the Azure Load Balancer with Standard SKU (hardcoded).
# Basic SKU was retired Sept 2025 and lacks zone redundancy, SLA, and outbound rules.
resource "azurerm_lb" "main" {
  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group
  sku                 = "Standard"
  tags                = local.final_tags

  frontend_ip_configuration {
    name                          = local.frontend_config_name
    public_ip_address_id          = local.is_internal ? null : var.spec.public_ip_id
    subnet_id                     = local.is_internal ? var.spec.subnet_id : null
    private_ip_address            = local.is_internal && var.spec.private_ip_address != null ? var.spec.private_ip_address : null
    private_ip_address_allocation = local.is_internal ? (var.spec.private_ip_address != null && var.spec.private_ip_address != "" ? "Static" : "Dynamic") : null
  }
}

# Create backend address pools.
# Pool membership (VMs, VMSS, NICs) is managed externally.
resource "azurerm_lb_backend_address_pool" "pools" {
  for_each = { for pool in var.spec.backend_pools : pool.name => pool }

  name            = each.value.name
  loadbalancer_id = azurerm_lb.main.id
}

# Create health probes.
resource "azurerm_lb_probe" "probes" {
  for_each = { for probe in var.spec.health_probes : probe.name => probe }

  name                = each.value.name
  loadbalancer_id     = azurerm_lb.main.id
  protocol            = each.value.protocol
  port                = each.value.port
  request_path        = each.value.protocol != "Tcp" ? each.value.request_path : null
  interval_in_seconds = each.value.interval_in_seconds
  number_of_probes    = each.value.number_of_probes
}

# Create load balancing rules.
# Each rule references a backend pool and probe by name.
resource "azurerm_lb_rule" "rules" {
  for_each = { for rule in var.spec.rules : rule.name => rule }

  name                           = each.value.name
  loadbalancer_id                = azurerm_lb.main.id
  frontend_ip_configuration_name = local.frontend_config_name
  protocol                       = each.value.protocol
  frontend_port                  = each.value.frontend_port
  backend_port                   = each.value.backend_port
  backend_address_pool_ids       = [local.backend_pool_map[each.value.backend_pool_name]]
  probe_id                       = local.probe_map[each.value.probe_name]
  idle_timeout_in_minutes        = each.value.idle_timeout_in_minutes
  enable_floating_ip             = each.value.enable_floating_ip
  disable_outbound_snat          = each.value.disable_outbound_snat
}
