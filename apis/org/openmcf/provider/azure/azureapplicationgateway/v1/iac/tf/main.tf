resource "azurerm_application_gateway" "main" {
  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group
  tags                = local.final_tags
  enable_http2        = var.spec.enable_http2

  sku {
    name     = var.spec.sku
    tier     = var.spec.sku
    capacity = local.use_autoscale ? null : var.spec.capacity
  }

  dynamic "autoscale_configuration" {
    for_each = local.use_autoscale ? [var.spec.autoscale] : []
    content {
      min_capacity = autoscale_configuration.value.min_capacity
      max_capacity = autoscale_configuration.value.max_capacity
    }
  }

  gateway_ip_configuration {
    name      = local.gateway_ip_config_name
    subnet_id = var.spec.subnet_id
  }

  frontend_ip_configuration {
    name                 = local.frontend_ip_config_name
    public_ip_address_id = var.spec.public_ip_id
  }

  dynamic "frontend_port" {
    for_each = local.frontend_ports
    content {
      name = frontend_port.key
      port = frontend_port.value
    }
  }

  dynamic "backend_address_pool" {
    for_each = { for pool in var.spec.backend_address_pools : pool.name => pool }
    content {
      name         = backend_address_pool.value.name
      fqdns        = length(backend_address_pool.value.fqdns) > 0 ? backend_address_pool.value.fqdns : null
      ip_addresses = length(backend_address_pool.value.ip_addresses) > 0 ? backend_address_pool.value.ip_addresses : null
    }
  }

  dynamic "backend_http_settings" {
    for_each = { for settings in var.spec.backend_http_settings : settings.name => settings }
    content {
      name                                = backend_http_settings.value.name
      port                                = backend_http_settings.value.port
      protocol                            = backend_http_settings.value.protocol
      cookie_based_affinity               = backend_http_settings.value.cookie_based_affinity
      request_timeout                     = backend_http_settings.value.request_timeout
      probe_name                          = backend_http_settings.value.probe_name
      host_name                           = backend_http_settings.value.host_name
      pick_host_name_from_backend_address = backend_http_settings.value.pick_host_name_from_backend_address
    }
  }

  dynamic "http_listener" {
    for_each = { for listener in var.spec.http_listeners : listener.name => listener }
    content {
      name                           = http_listener.value.name
      frontend_ip_configuration_name = local.frontend_ip_config_name
      frontend_port_name             = "${http_listener.value.name}-port"
      protocol                       = http_listener.value.protocol
      host_name                      = http_listener.value.host_name
      ssl_certificate_name           = http_listener.value.ssl_certificate_name
    }
  }

  dynamic "request_routing_rule" {
    for_each = { for rule in var.spec.request_routing_rules : rule.name => rule }
    content {
      name                       = request_routing_rule.value.name
      rule_type                  = "Basic"
      http_listener_name         = request_routing_rule.value.http_listener_name
      backend_address_pool_name  = request_routing_rule.value.backend_address_pool_name
      backend_http_settings_name = request_routing_rule.value.backend_http_settings_name
      priority                   = request_routing_rule.value.priority
    }
  }

  dynamic "probe" {
    for_each = { for p in var.spec.probes : p.name => p }
    content {
      name                = probe.value.name
      protocol            = probe.value.protocol
      path                = probe.value.path
      host                = probe.value.host
      interval            = probe.value.interval
      timeout             = probe.value.timeout
      unhealthy_threshold = probe.value.unhealthy_threshold
    }
  }

  dynamic "ssl_certificate" {
    for_each = { for cert in var.spec.ssl_certificates : cert.name => cert }
    content {
      name                = ssl_certificate.value.name
      key_vault_secret_id = ssl_certificate.value.key_vault_secret_id
    }
  }

  dynamic "identity" {
    for_each = length(var.spec.identity_ids) > 0 ? [1] : []
    content {
      type         = "UserAssigned"
      identity_ids = var.spec.identity_ids
    }
  }

  dynamic "waf_configuration" {
    for_each = var.spec.waf_enabled ? [1] : []
    content {
      enabled          = true
      firewall_mode    = var.spec.waf_mode
      rule_set_type    = "OWASP"
      rule_set_version = "3.2"
    }
  }
}
