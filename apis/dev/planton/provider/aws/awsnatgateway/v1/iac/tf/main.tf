resource "aws_nat_gateway" "this" {
  connectivity_type = var.spec.connectivity_type
  subnet_id         = var.spec.subnet_id

  # Public gateways: Elastic IP allocation(s).
  allocation_id            = local.allocation_id
  secondary_allocation_ids = local.secondary_allocation_ids

  # Private gateways: private IPv4 addressing.
  private_ip                         = local.private_ip
  secondary_private_ip_addresses     = local.secondary_private_ip_addresses
  secondary_private_ip_address_count = local.secondary_private_ip_address_count

  tags = local.aws_tags
}
