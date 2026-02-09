# main.tf

# Create the OpenStack Neutron subnet.
resource "openstack_networking_subnet_v2" "main" {
  name       = local.subnet_name
  network_id = local.network_id
  cidr       = var.spec.cidr
  ip_version = var.spec.ip_version

  # Description (empty means unset)
  description = var.spec.description != "" ? var.spec.description : null

  # Gateway configuration: gateway_ip and no_gateway are mutually exclusive.
  # - If no_gateway is true, disable the gateway entirely.
  # - If gateway_ip is set, use that specific IP.
  # - If neither is set, OpenStack auto-assigns the first usable IP.
  no_gateway = var.spec.no_gateway ? true : null
  gateway_ip = var.spec.no_gateway ? null : (var.spec.gateway_ip != "" ? var.spec.gateway_ip : null)

  # DHCP configuration
  enable_dhcp = var.spec.enable_dhcp

  # DNS nameservers (empty list means unset)
  dns_nameservers = length(var.spec.dns_nameservers) > 0 ? var.spec.dns_nameservers : null

  # Allocation pools
  dynamic "allocation_pool" {
    for_each = var.spec.allocation_pools
    content {
      start = allocation_pool.value.start
      end   = allocation_pool.value.end
    }
  }

  # Tags applied to the OpenStack resource
  tags = length(var.spec.tags) > 0 ? toset(var.spec.tags) : null

  # Region override (optional)
  region = var.spec.region != "" ? var.spec.region : null
}
