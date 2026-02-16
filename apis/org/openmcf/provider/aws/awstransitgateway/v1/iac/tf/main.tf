resource "aws_ec2_transit_gateway" "this" {
  description = var.spec.description

  amazon_side_asn = var.spec.amazon_side_asn

  auto_accept_shared_attachments     = local.enable_disable[var.spec.auto_accept_shared_attachments]
  default_route_table_association    = local.enable_disable[var.spec.default_route_table_association]
  default_route_table_propagation    = local.enable_disable[var.spec.default_route_table_propagation]
  dns_support                        = local.enable_disable[var.spec.dns_support]
  vpn_ecmp_support                   = local.enable_disable[var.spec.vpn_ecmp_support]
  security_group_referencing_support = local.enable_disable[var.spec.security_group_referencing_support]
  multicast_support                  = local.enable_disable[var.spec.multicast_support]

  transit_gateway_cidr_blocks = var.spec.transit_gateway_cidr_blocks

  tags = merge(local.tags, {
    Name = local.name
  })
}

resource "aws_ec2_transit_gateway_vpc_attachment" "this" {
  for_each = local.vpc_attachments_map

  transit_gateway_id = aws_ec2_transit_gateway.this.id
  vpc_id             = each.value.vpc_id
  subnet_ids         = each.value.subnet_ids

  dns_support          = local.enable_disable[each.value.dns_support]
  ipv6_support         = local.enable_disable[each.value.ipv6_support]
  appliance_mode_support = local.enable_disable[each.value.appliance_mode_support]

  transit_gateway_default_route_table_association = each.value.default_route_table_association
  transit_gateway_default_route_table_propagation = each.value.default_route_table_propagation

  tags = merge(local.tags, {
    Name = "${local.name}-${each.key}"
  })

  depends_on = [aws_ec2_transit_gateway.this]
}
