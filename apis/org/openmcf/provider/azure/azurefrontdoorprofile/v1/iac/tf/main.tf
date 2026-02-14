# Create the Azure Front Door profile.
resource "azurerm_cdn_frontdoor_profile" "main" {
  name                     = var.spec.name
  resource_group_name      = var.spec.resource_group
  sku_name                 = var.spec.sku
  response_timeout_seconds = var.spec.response_timeout_seconds

  tags = local.final_tags
}

# Create endpoints within the profile.
resource "azurerm_cdn_frontdoor_endpoint" "endpoints" {
  for_each = { for ep in var.spec.endpoints : ep.name => ep }

  name                     = each.value.name
  cdn_frontdoor_profile_id = azurerm_cdn_frontdoor_profile.main.id
  enabled                  = each.value.enabled
}

# Create origin groups within the profile.
resource "azurerm_cdn_frontdoor_origin_group" "origin_groups" {
  for_each = { for og in var.spec.origin_groups : og.name => og }

  name                     = each.value.name
  cdn_frontdoor_profile_id = azurerm_cdn_frontdoor_profile.main.id
  session_affinity_enabled = each.value.session_affinity_enabled

  load_balancing {
    sample_size                        = each.value.load_balancing.sample_size
    successful_samples_required        = each.value.load_balancing.successful_samples_required
    additional_latency_in_milliseconds = each.value.load_balancing.additional_latency_in_milliseconds
  }

  dynamic "health_probe" {
    for_each = each.value.health_probe != null ? [each.value.health_probe] : []

    content {
      protocol            = health_probe.value.protocol
      path                = health_probe.value.path
      request_type        = health_probe.value.request_type
      interval_in_seconds = health_probe.value.interval_in_seconds
    }
  }
}

# Create origins within origin groups.
resource "azurerm_cdn_frontdoor_origin" "origins" {
  for_each = local.origins_flat

  name                           = each.value.name
  cdn_frontdoor_origin_group_id  = azurerm_cdn_frontdoor_origin_group.origin_groups[each.value.origin_group_name].id
  host_name                      = each.value.host_name
  certificate_name_check_enabled = each.value.certificate_name_check_enabled
  origin_host_header             = each.value.origin_host_header
  http_port                      = each.value.http_port
  https_port                     = each.value.https_port
  priority                       = each.value.priority
  weight                         = each.value.weight
  enabled                        = each.value.enabled

  dynamic "private_link" {
    for_each = each.value.private_link != null ? [each.value.private_link] : []

    content {
      location               = private_link.value.location
      private_link_target_id = private_link.value.private_link_target_id
      request_message        = private_link.value.request_message
      target_type            = private_link.value.target_type
    }
  }
}

# Create routes connecting endpoints to origin groups.
resource "azurerm_cdn_frontdoor_route" "routes" {
  for_each = { for rt in var.spec.routes : rt.name => rt }

  name                          = each.value.name
  cdn_frontdoor_endpoint_id     = azurerm_cdn_frontdoor_endpoint.endpoints[each.value.endpoint_name].id
  cdn_frontdoor_origin_group_id = azurerm_cdn_frontdoor_origin_group.origin_groups[each.value.origin_group_name].id
  cdn_frontdoor_origin_ids = [
    for key, origin in azurerm_cdn_frontdoor_origin.origins :
    origin.id if startswith(key, "${each.value.origin_group_name}/")
  ]

  patterns_to_match      = each.value.patterns_to_match
  supported_protocols    = each.value.supported_protocols
  forwarding_protocol    = each.value.forwarding_protocol
  https_redirect_enabled = each.value.https_redirect_enabled
  link_to_default_domain = each.value.link_to_default_domain
  enabled                = each.value.enabled

  dynamic "cache" {
    for_each = each.value.cache != null ? [each.value.cache] : []

    content {
      query_string_caching_behavior = cache.value.query_string_caching_behavior
      query_strings                 = cache.value.query_strings
      compression_enabled           = cache.value.compression_enabled
      content_types_to_compress     = cache.value.content_types_to_compress
    }
  }
}
