locals {
  endpoint_name = var.metadata.name
}

resource "aws_security_group" "this" {
  count = length(var.spec.security_groups) == 0 ? 1 : 0

  name        = "${local.endpoint_name}-cvpn"
  description = "Security group for Client VPN endpoint"
  vpc_id      = var.spec.vpc_id.value
}

resource "aws_ec2_client_vpn_endpoint" "this" {
  description       = try(var.spec.description, null)
  server_certificate_arn = var.spec.server_certificate_arn.value
  client_cidr_block      = var.spec.client_cidr_block
  split_tunnel           = !try(var.spec.disable_split_tunnel, false)
  transport_protocol     = try(var.spec.transport_protocol, "tcp")
  vpn_port               = try(var.spec.vpn_port, 443)

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = var.spec.server_certificate_arn.value
  }

  connection_log_options {
    enabled              = try(var.spec.log_group_name, "") != ""
    cloudwatch_log_group = try(var.spec.log_group_name, null)
  }

  dns_servers = length(local.safe_dns_servers) > 0 ? local.safe_dns_servers : null

  # Apply security groups to target network associations. If none provided, use the module-created SG.
  security_group_ids = local.association_security_group_ids

  tags = {
    Name = local.endpoint_name
  }
}

locals {
  association_security_group_ids = length(var.spec.security_groups) == 0 ? [aws_security_group.this[0].id] : [for sg in var.spec.security_groups : sg.value]
}

resource "aws_ec2_client_vpn_network_association" "this" {
  for_each = { for s in var.spec.subnets : s.value => s }

  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.this.id
  subnet_id              = each.value.value
}

resource "aws_ec2_client_vpn_authorization_rule" "cidr" {
  for_each = { for c in local.safe_cidr_authorization_rules : c => c }

  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.this.id
  target_network_cidr    = each.value
  authorize_all_groups   = true
}


