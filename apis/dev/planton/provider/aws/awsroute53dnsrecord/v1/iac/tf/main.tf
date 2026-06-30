# AWS Route53 DNS Record
resource "aws_route53_record" "record" {
  zone_id = local.zone_id
  name    = var.spec.name
  type    = var.spec.type

  # TTL and records for standard (non-alias) records
  ttl     = local.ttl
  records = local.records

  # Set identifier for routing policies
  set_identifier = var.spec.set_identifier

  # Health check for failover routing
  health_check_id = var.spec.health_check_id

  # Alias record configuration
  dynamic "alias" {
    for_each = local.is_alias ? [1] : []
    content {
      name                   = local.alias_dns_name
      zone_id                = local.alias_zone_id
      evaluate_target_health = var.spec.alias_target.evaluate_target_health
    }
  }

  # Weighted routing policy
  dynamic "weighted_routing_policy" {
    for_each = local.has_weighted ? [var.spec.routing_policy.weighted] : []
    content {
      weight = weighted_routing_policy.value.weight
    }
  }

  # Latency-based routing policy
  dynamic "latency_routing_policy" {
    for_each = local.has_latency ? [var.spec.routing_policy.latency] : []
    content {
      region = latency_routing_policy.value.region
    }
  }

  # Failover routing policy
  dynamic "failover_routing_policy" {
    for_each = local.has_failover ? [var.spec.routing_policy.failover] : []
    content {
      type = upper(failover_routing_policy.value.failover_type)
    }
  }

  # Geolocation routing policy
  dynamic "geolocation_routing_policy" {
    for_each = local.has_geolocation ? [var.spec.routing_policy.geolocation] : []
    content {
      continent   = geolocation_routing_policy.value.continent
      country     = geolocation_routing_policy.value.country
      subdivision = geolocation_routing_policy.value.subdivision
    }
  }
}
