# Create Virtual Network
resource "azurerm_virtual_network" "main" {
  name                = local.vnet_name
  resource_group_name = local.resource_group_name
  location            = var.spec.region
  address_space       = [var.spec.address_space_cidr]
  tags                = local.final_tags
}

# Create Public IP for NAT Gateway (if NAT Gateway is enabled)
resource "azurerm_public_ip" "nat" {
  count = var.spec.is_nat_gateway_enabled ? 1 : 0

  name                = local.public_ip_name
  resource_group_name = local.resource_group_name
  location            = var.spec.region
  allocation_method   = "Static"
  sku                 = "Standard"
  tags                = local.final_tags
}

# Create NAT Gateway (if enabled)
resource "azurerm_nat_gateway" "main" {
  count = var.spec.is_nat_gateway_enabled ? 1 : 0

  name                    = local.nat_gateway_name
  resource_group_name     = local.resource_group_name
  location                = var.spec.region
  sku_name                = "Standard"
  idle_timeout_in_minutes = 4
  tags                    = local.final_tags
}

# Associate Public IP with NAT Gateway
resource "azurerm_nat_gateway_public_ip_association" "main" {
  count = var.spec.is_nat_gateway_enabled ? 1 : 0

  nat_gateway_id       = azurerm_nat_gateway.main[0].id
  public_ip_address_id = azurerm_public_ip.nat[0].id
}

# Create Subnet for nodes
resource "azurerm_subnet" "nodes" {
  name                 = local.subnet_name
  resource_group_name  = local.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.spec.nodes_subnet_cidr]

  depends_on = [
    azurerm_virtual_network.main
  ]
}

# Associate NAT Gateway with Subnet (if enabled)
resource "azurerm_subnet_nat_gateway_association" "main" {
  count = var.spec.is_nat_gateway_enabled ? 1 : 0

  subnet_id      = azurerm_subnet.nodes.id
  nat_gateway_id = azurerm_nat_gateway.main[0].id

  depends_on = [
    azurerm_nat_gateway_public_ip_association.main
  ]
}

# Link Private DNS Zones to VNet (if specified)
resource "azurerm_private_dns_zone_virtual_network_link" "dns_links" {
  for_each = toset(var.spec.dns_private_zone_links)

  name                  = "${local.vnet_name}-link-${index(var.spec.dns_private_zone_links, each.value)}"
  resource_group_name   = local.resource_group_name
  private_dns_zone_name = element(reverse(split("/", each.value)), 0)
  virtual_network_id    = azurerm_virtual_network.main.id
  registration_enabled  = false
  tags                  = local.final_tags
}

