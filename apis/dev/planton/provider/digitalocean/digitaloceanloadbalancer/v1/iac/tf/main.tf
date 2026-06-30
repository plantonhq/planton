# DigitalOcean Load Balancer Resource
# This module creates a regional load balancer in DigitalOcean with the specified configuration

resource "digitalocean_loadbalancer" "this" {
  name   = var.spec.load_balancer_name
  region = var.spec.region

  # VPC placement - load balancer communicates with Droplets over private network
  vpc_uuid = local.vpc_uuid

  # Backend targeting - either tag-based (dynamic) or ID-based (static)
  # Tag-based is recommended for production (enables autoscaling and blue-green deployments)
  droplet_tag = var.spec.droplet_tag
  droplet_ids = length(local.valid_droplet_ids) > 0 ? local.valid_droplet_ids : null

  # Forwarding rules - how traffic is routed from load balancer to backends
  dynamic "forwarding_rule" {
    for_each = local.forwarding_rules
    content {
      entry_port      = forwarding_rule.value.entry_port
      entry_protocol  = forwarding_rule.value.entry_protocol
      target_port     = forwarding_rule.value.target_port
      target_protocol = forwarding_rule.value.target_protocol

      # SSL certificate for HTTPS termination (use certificate name, not ID)
      certificate_name = forwarding_rule.value.certificate_name != null && forwarding_rule.value.certificate_name != "" ? forwarding_rule.value.certificate_name : null
    }
  }

  # Health check configuration - determines which backends are healthy
  dynamic "healthcheck" {
    for_each = local.healthcheck != null ? [local.healthcheck] : []
    content {
      port                   = healthcheck.value.port
      protocol               = healthcheck.value.protocol
      path                   = healthcheck.value.path != null && healthcheck.value.path != "" ? healthcheck.value.path : null
      check_interval_seconds = healthcheck.value.check_interval_seconds
    }
  }

  # Sticky sessions configuration (optional)
  dynamic "sticky_sessions" {
    for_each = local.sticky_sessions != null ? [local.sticky_sessions] : []
    content {
      type = sticky_sessions.value.type
    }
  }

  # Lifecycle management
  lifecycle {
    # Prevent destruction if load balancer has active traffic
    # Remove this if you need to force replacement
    prevent_destroy = false

    # Ignore changes to droplet_ids if using tag-based targeting
    # This prevents Terraform from constantly trying to update the droplet list
    ignore_changes = [
      droplet_ids
    ]
  }
}

