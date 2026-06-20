# A subnet-owned route table is created only when inline routes are supplied.
# Each route maps its destination and its typed target onto the matching AWS
# route attributes (others left null).
resource "aws_route_table" "this" {
  count  = length(var.spec.routes) > 0 ? 1 : 0
  vpc_id = var.spec.vpc_id

  dynamic "route" {
    for_each = var.spec.routes
    content {
      cidr_block                 = route.value.destination_cidr_block != "" ? route.value.destination_cidr_block : null
      ipv6_cidr_block            = route.value.destination_ipv6_cidr_block != "" ? route.value.destination_ipv6_cidr_block : null
      destination_prefix_list_id = route.value.destination_prefix_list_id != "" ? route.value.destination_prefix_list_id : null

      gateway_id                = route.value.target_type == "internet_gateway" ? route.value.target_id : null
      nat_gateway_id            = route.value.target_type == "nat_gateway" ? route.value.target_id : null
      transit_gateway_id        = route.value.target_type == "transit_gateway" ? route.value.target_id : null
      vpc_peering_connection_id = route.value.target_type == "vpc_peering_connection" ? route.value.target_id : null
      vpc_endpoint_id           = route.value.target_type == "vpc_endpoint" ? route.value.target_id : null
      network_interface_id      = route.value.target_type == "network_interface" ? route.value.target_id : null
      egress_only_gateway_id    = route.value.target_type == "egress_only_internet_gateway" ? route.value.target_id : null
    }
  }

  tags = local.aws_tags
}

# Associate the subnet with its inline-created table, or with the externally
# referenced route_table_id. When neither is set, the subnet stays on the VPC
# main route table and no association is created.
resource "aws_route_table_association" "this" {
  count = (length(var.spec.routes) > 0 || var.spec.route_table_id != "") ? 1 : 0

  subnet_id = aws_subnet.this.id
  route_table_id = (
    length(var.spec.routes) > 0
    ? aws_route_table.this[0].id
    : var.spec.route_table_id
  )
}
