# ──────────────────────────────────────────────────────────────────────────────
# Network Load Balancer
# ──────────────────────────────────────────────────────────────────────────────

resource "aws_lb" "this" {
  name               = var.metadata.name
  load_balancer_type = "network"
  internal           = var.spec.internal
  ip_address_type    = local.ip_address_type

  enable_deletion_protection    = var.spec.delete_protection_enabled
  enable_cross_zone_load_balancing = var.spec.cross_zone_load_balancing_enabled

  dns_record_client_routing_policy = var.spec.dns_record_client_routing_policy != "" ? var.spec.dns_record_client_routing_policy : null

  dynamic "subnet_mapping" {
    for_each = var.spec.subnet_mappings
    content {
      subnet_id            = subnet_mapping.value.subnet_id
      allocation_id        = lookup(subnet_mapping.value, "allocation_id", null)
      private_ipv4_address = lookup(subnet_mapping.value, "private_ipv4_address", null)
    }
  }

  security_groups = length(var.spec.security_groups) > 0 ? var.spec.security_groups : null

  tags = local.tags
}

# ──────────────────────────────────────────────────────────────────────────────
# Target Groups (one per listener)
# ──────────────────────────────────────────────────────────────────────────────

resource "aws_lb_target_group" "this" {
  for_each = local.listener_map

  name        = substr("${var.metadata.name}-${each.key}", 0, 32)
  port        = each.value.target_group.port
  protocol    = each.value.target_group.protocol
  vpc_id      = data.aws_lb.this_vpc.vpc_id
  target_type = coalesce(lookup(each.value.target_group, "target_type", null), "instance")

  deregistration_delay = lookup(each.value.target_group, "deregistration_delay_seconds", null)

  preserve_client_ip   = lookup(each.value.target_group, "preserve_client_ip", null) ? "true" : null
  proxy_protocol_v2    = lookup(each.value.target_group, "proxy_protocol_v2", false)
  connection_termination = lookup(each.value.target_group, "connection_termination", false)

  dynamic "stickiness" {
    for_each = lookup(each.value.target_group, "stickiness_enabled", false) ? [1] : []
    content {
      enabled = true
      type    = "source_ip"
    }
  }

  dynamic "health_check" {
    for_each = lookup(each.value.target_group, "health_check", null) != null ? [each.value.target_group.health_check] : []
    content {
      protocol            = lookup(health_check.value, "protocol", null)
      port                = lookup(health_check.value, "port", null)
      path                = lookup(health_check.value, "path", null)
      healthy_threshold   = lookup(health_check.value, "healthy_threshold", null)
      unhealthy_threshold = lookup(health_check.value, "unhealthy_threshold", null)
      interval            = lookup(health_check.value, "interval_seconds", null)
      timeout             = lookup(health_check.value, "timeout_seconds", null)
      matcher             = lookup(health_check.value, "matcher", null)
    }
  }

  tags = local.tags
}

# ──────────────────────────────────────────────────────────────────────────────
# Listeners (one per spec listener)
# ──────────────────────────────────────────────────────────────────────────────

resource "aws_lb_listener" "this" {
  for_each = local.listener_map

  load_balancer_arn = aws_lb.this.arn
  port              = each.value.port
  protocol          = each.value.protocol

  certificate_arn          = lookup(each.value, "tls", null) != null ? each.value.tls.certificate_arn : null
  ssl_policy               = lookup(each.value, "tls", null) != null ? lookup(each.value.tls, "ssl_policy", null) : null
  alpn_policy              = lookup(each.value, "alpn_policy", null) != "" ? lookup(each.value, "alpn_policy", null) : null
  tcp_idle_timeout_seconds = lookup(each.value, "tcp_idle_timeout_seconds", null) > 0 ? each.value.tcp_idle_timeout_seconds : null

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.this[each.key].arn
  }

  tags = local.tags
}

# ──────────────────────────────────────────────────────────────────────────────
# Data source: resolve VPC ID from the NLB (after creation)
# ──────────────────────────────────────────────────────────────────────────────

data "aws_lb" "this_vpc" {
  depends_on = [aws_lb.this]
  arn        = aws_lb.this.arn
}

# ──────────────────────────────────────────────────────────────────────────────
# Route53 DNS Records (optional)
# ──────────────────────────────────────────────────────────────────────────────

# allow_overwrite adopts an existing alias record (e.g. left by a prior partial apply,
# or one already pointing at this NLB) instead of failing the apply on a CREATE
# collision -- this alias record is owned by the NLB module.
resource "aws_route53_record" "this" {
  for_each = local.dns_records

  allow_overwrite = true
  zone_id         = var.spec.dns.route53_zone_id
  name            = each.value
  type            = "A"

  alias {
    name                   = aws_lb.this.dns_name
    zone_id                = aws_lb.this.zone_id
    evaluate_target_health = true
  }
}
