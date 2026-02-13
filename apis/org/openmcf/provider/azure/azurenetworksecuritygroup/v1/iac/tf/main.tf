# Create the Azure Network Security Group.
resource "azurerm_network_security_group" "main" {
  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group
  tags                = local.final_tags
}

# Create individual security rules as separate resources.
# Using separate resources (not inline) gives each rule its own lifecycle
# and avoids the Terraform "inline vs separate" conflict issue.
resource "azurerm_network_security_rule" "rules" {
  for_each = { for rule in var.spec.security_rules : rule.name => rule }

  name                        = each.value.name
  resource_group_name         = var.spec.resource_group
  network_security_group_name = azurerm_network_security_group.main.name

  description = each.value.description
  priority    = each.value.priority
  direction   = each.value.direction
  access      = each.value.access
  protocol    = each.value.protocol

  source_port_range      = each.value.source_port_range
  destination_port_range = each.value.destination_port_range

  # Address prefix precedence: plural overrides singular if non-empty.
  source_address_prefix    = length(coalesce(each.value.source_address_prefixes, [])) > 0 ? null : each.value.source_address_prefix
  source_address_prefixes  = length(coalesce(each.value.source_address_prefixes, [])) > 0 ? each.value.source_address_prefixes : null

  destination_address_prefix   = length(coalesce(each.value.destination_address_prefixes, [])) > 0 ? null : each.value.destination_address_prefix
  destination_address_prefixes = length(coalesce(each.value.destination_address_prefixes, [])) > 0 ? each.value.destination_address_prefixes : null
}
