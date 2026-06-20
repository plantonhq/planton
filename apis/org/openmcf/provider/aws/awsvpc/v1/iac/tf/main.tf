resource "aws_vpc" "this" {
  cidr_block          = var.spec.cidr_block != "" ? var.spec.cidr_block : null
  ipv4_ipam_pool_id   = var.spec.ipv4_ipam_pool_id != "" ? var.spec.ipv4_ipam_pool_id : null
  ipv4_netmask_length = var.spec.ipv4_netmask_length != 0 ? var.spec.ipv4_netmask_length : null

  instance_tenancy                     = var.spec.instance_tenancy != "" ? var.spec.instance_tenancy : null
  enable_dns_support                   = var.spec.enable_dns_support
  enable_dns_hostnames                 = var.spec.enable_dns_hostnames
  enable_network_address_usage_metrics = var.spec.enable_network_address_usage_metrics

  assign_generated_ipv6_cidr_block     = var.spec.assign_generated_ipv6_cidr_block
  ipv6_cidr_block                      = var.spec.ipv6_cidr_block != "" ? var.spec.ipv6_cidr_block : null
  ipv6_cidr_block_network_border_group = var.spec.ipv6_cidr_block_network_border_group != "" ? var.spec.ipv6_cidr_block_network_border_group : null
  ipv6_ipam_pool_id                    = var.spec.ipv6_ipam_pool_id != "" ? var.spec.ipv6_ipam_pool_id : null
  ipv6_netmask_length                  = var.spec.ipv6_netmask_length != 0 ? var.spec.ipv6_netmask_length : null

  tags = local.aws_tags
}

# Each secondary IPv4 CIDR is its own association so it can be added or removed
# without recreating the VPC.
resource "aws_vpc_ipv4_cidr_block_association" "secondary" {
  for_each = toset(var.spec.secondary_ipv4_cidr_blocks)

  vpc_id     = aws_vpc.this.id
  cidr_block = each.value
}
