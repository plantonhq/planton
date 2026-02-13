# Scaleway Private Network Resource
resource "scaleway_vpc_private_network" "private_network" {
  name   = local.private_network_name
  vpc_id = local.vpc_id
  region = local.region
  tags   = local.standard_tags

  # IPv4 subnet (optional).
  # If not specified, Scaleway's IPAM automatically allocates a subnet.
  dynamic "ipv4_subnet" {
    for_each = local.ipv4_subnet != "" ? [local.ipv4_subnet] : []
    content {
      subnet = ipv4_subnet.value
    }
  }

  # IPv6 subnets (optional).
  # Each entry creates a separate IPv6 subnet block.
  dynamic "ipv6_subnets" {
    for_each = local.ipv6_subnets
    content {
      subnet = ipv6_subnets.value
    }
  }

  # Default route propagation.
  # When enabled, resources in this network receive the VPC's default routes.
  enable_default_route_propagation = local.enable_default_route_propagation
}
