resource "aws_egress_only_internet_gateway" "this" {
  vpc_id = var.spec.vpc_id

  tags = local.aws_tags
}
