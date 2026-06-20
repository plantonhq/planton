resource "aws_subnet" "this" {
  vpc_id            = var.spec.vpc_id
  cidr_block        = var.spec.cidr_block
  availability_zone = var.spec.availability_zone

  map_public_ip_on_launch                        = var.spec.map_public_ip_on_launch
  assign_ipv6_address_on_creation                = var.spec.assign_ipv6_address_on_creation
  enable_dns64                                   = var.spec.enable_dns64
  enable_resource_name_dns_a_record_on_launch    = var.spec.enable_resource_name_dns_a_record_on_launch
  enable_resource_name_dns_aaaa_record_on_launch = var.spec.enable_resource_name_dns_aaaa_record_on_launch

  ipv6_cidr_block                     = var.spec.ipv6_cidr_block != "" ? var.spec.ipv6_cidr_block : null
  private_dns_hostname_type_on_launch = var.spec.private_dns_hostname_type_on_launch != "" ? var.spec.private_dns_hostname_type_on_launch : null

  tags = local.aws_tags
}
