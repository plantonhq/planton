# Security Group for DocumentDB Cluster
# Created when security_group_ids or allowed_cidr_blocks are provided

resource "aws_security_group" "main" {
  count = local.need_managed_sg ? 1 : 0

  name        = local.resource_id
  description = "Security group for DocumentDB cluster ${local.resource_id}"
  vpc_id      = local.vpc_id

  tags = local.final_tags
}

# Ingress rules from security groups
resource "aws_security_group_rule" "ingress_from_sg" {
  count = local.need_managed_sg ? length(local.ingress_sg_ids) : 0

  type                     = "ingress"
  from_port                = local.port
  to_port                  = local.port
  protocol                 = "tcp"
  source_security_group_id = local.ingress_sg_ids[count.index]
  security_group_id        = aws_security_group.main[0].id
}

# Ingress rules from CIDR blocks
resource "aws_security_group_rule" "ingress_from_cidr" {
  count = local.need_managed_sg && length(local.allowed_cidrs) > 0 ? 1 : 0

  type              = "ingress"
  from_port         = local.port
  to_port           = local.port
  protocol          = "tcp"
  cidr_blocks       = local.allowed_cidrs
  security_group_id = aws_security_group.main[0].id
}

# Egress rule - allow all outbound
resource "aws_security_group_rule" "egress_all" {
  count = local.need_managed_sg ? 1 : 0

  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.main[0].id
}
