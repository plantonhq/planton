# Create Public IP Prefix (if prefix length specified)
resource "azurerm_public_ip_prefix" "nat" {
  count = local.use_ip_prefix ? 1 : 0

  name                = "${local.nat_gateway_name}-prefix"
  resource_group_name = local.resource_group
  location            = local.location
  prefix_length       = var.spec.public_ip_prefix_length
  sku                 = "Standard"

  tags = local.final_tags
}

# Create individual Public IP (if no prefix specified)
resource "azurerm_public_ip" "nat" {
  count = local.use_ip_prefix ? 0 : 1

  name                = "${local.nat_gateway_name}-ip"
  resource_group_name = local.resource_group
  location            = local.location
  allocation_method   = "Static"
  sku                 = "Standard"

  tags = local.final_tags
}

# Create NAT Gateway
resource "azurerm_nat_gateway" "main" {
  name                    = local.nat_gateway_name
  resource_group_name     = local.resource_group
  location                = local.location
  sku_name                = "Standard"
  idle_timeout_in_minutes = var.spec.idle_timeout_minutes

  tags = local.final_tags
}

# Associate Public IP Prefix with NAT Gateway
resource "azurerm_nat_gateway_public_ip_prefix_association" "main" {
  count = local.use_ip_prefix ? 1 : 0

  nat_gateway_id      = azurerm_nat_gateway.main.id
  public_ip_prefix_id = azurerm_public_ip_prefix.nat[0].id
}

# Associate individual Public IP with NAT Gateway
resource "azurerm_nat_gateway_public_ip_association" "main" {
  count = local.use_ip_prefix ? 0 : 1

  nat_gateway_id       = azurerm_nat_gateway.main.id
  public_ip_address_id = azurerm_public_ip.nat[0].id
}

# Associate NAT Gateway with Subnet
resource "azurerm_subnet_nat_gateway_association" "main" {
  subnet_id      = var.spec.subnet_id
  nat_gateway_id = azurerm_nat_gateway.main.id

  depends_on = [
    azurerm_nat_gateway_public_ip_prefix_association.main,
    azurerm_nat_gateway_public_ip_association.main
  ]
}

