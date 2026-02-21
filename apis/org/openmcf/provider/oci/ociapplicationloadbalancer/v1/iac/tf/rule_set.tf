resource "oci_load_balancer_rule_set" "this" {
  for_each = local.rule_sets_map

  load_balancer_id = oci_load_balancer_load_balancer.this.id
  name             = each.value.name

  dynamic "items" {
    for_each = each.value.items
    content {
      action = lookup(local.action_map, items.value.action, upper(items.value.action))

      header      = items.value.header != "" ? items.value.header : null
      value       = items.value.value != "" ? items.value.value : null
      prefix      = items.value.prefix != "" ? items.value.prefix : null
      suffix      = items.value.suffix != "" ? items.value.suffix : null
      description = items.value.description != "" ? items.value.description : null

      response_code = items.value.response_code > 0 ? items.value.response_code : null
      status_code   = items.value.status_code > 0 ? items.value.status_code : null

      allowed_methods              = length(items.value.allowed_methods) > 0 ? items.value.allowed_methods : null
      are_invalid_characters_allowed = items.value.are_invalid_characters_allowed ? true : null
      http_large_header_size_in_kb = items.value.http_large_header_size_in_kb > 0 ? items.value.http_large_header_size_in_kb : null
      default_max_connections      = items.value.default_max_connections > 0 ? items.value.default_max_connections : null

      dynamic "redirect_uri" {
        for_each = items.value.redirect_uri != null ? [items.value.redirect_uri] : []
        content {
          protocol = redirect_uri.value.protocol != "" ? redirect_uri.value.protocol : null
          host     = redirect_uri.value.host != "" ? redirect_uri.value.host : null
          port     = redirect_uri.value.port > 0 ? redirect_uri.value.port : null
          path     = redirect_uri.value.path != "" ? redirect_uri.value.path : null
          query    = redirect_uri.value.query != "" ? redirect_uri.value.query : null
        }
      }

      dynamic "conditions" {
        for_each = items.value.conditions
        content {
          attribute_name  = conditions.value.attribute_name
          attribute_value = conditions.value.attribute_value
          operator        = conditions.value.operator != "" ? conditions.value.operator : null
        }
      }

      dynamic "ip_max_connections" {
        for_each = items.value.ip_max_connections
        content {
          ip_addresses    = ip_max_connections.value.ip_addresses
          max_connections = ip_max_connections.value.max_connections
        }
      }
    }
  }
}
