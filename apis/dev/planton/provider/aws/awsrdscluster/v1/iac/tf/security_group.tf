resource "aws_security_group" "cluster" {
  count       = local.need_managed_sg ? 1 : 0
  name        = local.resource_id
  description = "Ingress for RDS cluster"
  vpc_id      = local.vpc_id
  tags        = local.final_labels
}

resource "aws_security_group_rule" "ingress_from_sg" {
  for_each = { for idx, sg_id in local.ingress_sg_ids : idx => sg_id }
  type                     = "ingress"
  from_port                = try(var.spec.port, 0)
  to_port                  = try(var.spec.port, 0)
  protocol                 = "tcp"
  source_security_group_id = each.value
  security_group_id        = aws_security_group.cluster[0].id
}

resource "aws_security_group_rule" "ingress_from_cidr" {
  count            = length(local.allowed_cidr_blocks) > 0 ? 1 : 0
  type             = "ingress"
  from_port        = try(var.spec.port, 0)
  to_port          = try(var.spec.port, 0)
  protocol         = "tcp"
  cidr_blocks      = local.allowed_cidr_blocks
  security_group_id = aws_security_group.cluster[0].id
}

resource "aws_security_group_rule" "egress_all" {
  count             = local.need_managed_sg ? 1 : 0
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.cluster[0].id
}


