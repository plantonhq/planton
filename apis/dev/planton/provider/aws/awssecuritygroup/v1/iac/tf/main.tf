# main.tf

resource "aws_security_group" "this" {
  name        = local.security_group_name
  description = local.description
  vpc_id      = local.vpc_id

  # ingress rules
  # each item in var.ingress maps to one AWS ingress rule
  dynamic "ingress" {
    for_each = var.spec.ingress
    content {
      description      = ingress.value.description
      protocol         = ingress.value.protocol
      from_port        = ingress.value.from_port
      to_port          = ingress.value.to_port
      cidr_blocks      = ingress.value.ipv4_cidrs
      ipv6_cidr_blocks = ingress.value.ipv6_cidrs
      security_groups  = ingress.value.source_security_group_ids
      self             = ingress.value.self_reference
    }
  }

  # egress rules
  # each item in var.egress maps to one AWS egress rule
  dynamic "egress" {
    for_each = var.spec.egress
    content {
      description      = egress.value.description
      protocol         = egress.value.protocol
      from_port        = egress.value.from_port
      to_port          = egress.value.to_port
      cidr_blocks      = egress.value.ipv4_cidrs
      ipv6_cidr_blocks = egress.value.ipv6_cidrs
      security_groups  = egress.value.destination_security_group_ids
      self             = egress.value.self_reference
    }
  }

  tags = local.final_labels
}
