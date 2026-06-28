resource "aws_lb" "this" {
  name                       = local.resource_name
  load_balancer_type         = "application"
  security_groups            = local.security_group_ids
  subnets                    = local.subnet_ids
  internal                   = try(var.spec.internal, false)
  ip_address_type            = "ipv4"
  enable_deletion_protection = try(var.spec.delete_protection_enabled, false)
  idle_timeout               = try(var.spec.idle_timeout_seconds, 60)

  tags = local.tags
}

# HTTP listener that redirects to HTTPS when SSL is enabled
resource "aws_lb_listener" "http_redirect" {
  count = local.is_ssl_enabled ? 1 : 0

  load_balancer_arn = aws_lb.this.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

# HTTPS listener when SSL is enabled
resource "aws_lb_listener" "https" {
  count = local.is_ssl_enabled ? 1 : 0

  load_balancer_arn = aws_lb.this.arn
  port              = 443
  protocol          = "HTTPS"
  certificate_arn   = local.certificate_arn
  ssl_policy        = "ELBSecurityPolicy-2016-08"

  default_action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "OK"
      status_code  = "200"
    }
  }
}

# Plain HTTP listener when SSL is NOT enabled. Without this an HTTP-only ALB has no
# listener at all, so downstream services (e.g. an ECS service) have nothing to attach
# their host/path routing rules to. The default action mirrors the https listener's
# fixed-response 200; services add forwarding rules on top of it.
resource "aws_lb_listener" "http" {
  count = local.is_ssl_enabled ? 0 : 1

  load_balancer_arn = aws_lb.this.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "OK"
      status_code  = "200"
    }
  }
}

# Optional Route53 records for each hostname when DNS is enabled.
# allow_overwrite adopts an existing alias record (e.g. left by a prior partial apply,
# or one already pointing at this ALB) instead of failing the apply on a CREATE
# collision -- this alias record is owned by the ALB module.
resource "aws_route53_record" "this" {
  for_each = local.create_dns_records ? toset(var.spec.dns.hostnames) : []

  allow_overwrite = true
  zone_id         = local.route53_zone_id
  name            = each.value
  type            = "A"

  alias {
    name                   = aws_lb.this.dns_name
    zone_id                = aws_lb.this.zone_id
    evaluate_target_health = false
  }
}


