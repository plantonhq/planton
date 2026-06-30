# ── 1. Dedicated Flexible IP ──────────────────────────────────────────────────
#
# A dedicated public IPv4 address for the Load Balancer. Creating it as a
# separate resource gives explicit lifecycle control -- the IP survives LB
# replacement, preserving DNS records and firewall rules.
resource "scaleway_lb_ip" "ip" {
  tags = local.standard_tags
  zone = local.zone
}

# ── 2. Load Balancer ─────────────────────────────────────────────────────────
#
# The managed traffic distribution appliance. Attached to the Flexible IP
# and optionally to a Private Network for backend connectivity.
resource "scaleway_lb" "lb" {
  name    = local.lb_name
  type    = local.lb_type
  ip_ids  = [scaleway_lb_ip.ip.id]
  tags    = local.standard_tags
  zone    = local.zone

  description              = local.description != "" ? local.description : null
  ssl_compatibility_level  = local.ssl_compatibility_level != "" ? local.ssl_compatibility_level : null

  # Attach to Private Network if specified.
  dynamic "private_network" {
    for_each = local.private_network_id != "" ? [1] : []
    content {
      private_network_id = local.private_network_id
    }
  }
}

# ── 3. TLS Certificates ─────────────────────────────────────────────────────
#
# TLS certificates for HTTPS frontends. Each certificate is either Let's
# Encrypt (auto-provisioned) or custom (user-provided PEM chain).
# Created before frontends because frontends reference certificate IDs.
resource "scaleway_lb_certificate" "certs" {
  for_each = local.certificates_map

  lb_id = scaleway_lb.lb.id
  name  = each.value.name

  # Let's Encrypt auto-provisioned certificate.
  dynamic "letsencrypt" {
    for_each = each.value.letsencrypt != null ? [each.value.letsencrypt] : []
    content {
      common_name              = letsencrypt.value.common_name
      subject_alternative_name = letsencrypt.value.subject_alternative_names
    }
  }

  # Custom certificate (user-provided PEM).
  dynamic "custom_certificate" {
    for_each = each.value.custom_certificate != null ? [each.value.custom_certificate] : []
    content {
      certificate_chain = custom_certificate.value.certificate_chain
    }
  }
}

# ── 4. Backend Server Pools ──────────────────────────────────────────────────
#
# Each backend defines a named group of servers with health checks and
# load-balancing configuration. Created before frontends because frontends
# reference backend IDs.
resource "scaleway_lb_backend" "backends" {
  for_each = local.backends_map

  lb_id            = scaleway_lb.lb.id
  name             = each.value.name
  forward_port     = each.value.forward_port
  forward_protocol = each.value.forward_protocol
  server_ips       = each.value.server_ips

  forward_port_algorithm      = each.value.forward_port_algorithm
  sticky_sessions             = each.value.sticky_sessions
  sticky_sessions_cookie_name = each.value.sticky_sessions_cookie_name != "" ? each.value.sticky_sessions_cookie_name : null

  timeout_connect       = each.value.timeout_connect != "" ? each.value.timeout_connect : null
  timeout_server        = each.value.timeout_server != "" ? each.value.timeout_server : null
  on_marked_down_action = each.value.on_marked_down_action != "" ? each.value.on_marked_down_action : null
  ssl_bridging          = each.value.ssl_bridging
  proxy_protocol        = each.value.proxy_protocol

  # Health check configuration.
  health_check_delay       = each.value.health_check != null ? each.value.health_check.check_delay : "5s"
  health_check_timeout     = each.value.health_check != null ? each.value.health_check.check_timeout : "3s"
  health_check_max_retries = each.value.health_check != null ? each.value.health_check.check_max_retries : 3
  health_check_port        = each.value.health_check != null && each.value.health_check.port > 0 ? each.value.health_check.port : each.value.forward_port

  # HTTP health check (when type = "http").
  dynamic "health_check_http" {
    for_each = each.value.health_check != null && each.value.health_check.type == "http" ? [each.value.health_check] : []
    content {
      uri  = health_check_http.value.uri
      code = health_check_http.value.expected_code
    }
  }

  # HTTPS health check (when type = "https").
  dynamic "health_check_https" {
    for_each = each.value.health_check != null && each.value.health_check.type == "https" ? [each.value.health_check] : []
    content {
      uri  = health_check_https.value.uri
      code = health_check_https.value.expected_code
    }
  }

  # TCP health check (default, when type = "tcp" or unspecified).
  dynamic "health_check_tcp" {
    for_each = each.value.health_check == null || each.value.health_check.type == "tcp" ? [1] : []
    content {}
  }
}

# ── 5. Frontend Listeners ────────────────────────────────────────────────────
#
# Each frontend listens on a port and routes traffic to a backend. Frontends
# reference backends and certificates by name, resolved to IDs via for_each keys.
resource "scaleway_lb_frontend" "frontends" {
  for_each = local.frontends_map

  lb_id       = scaleway_lb.lb.id
  name        = each.value.name
  inbound_port = each.value.inbound_port
  backend_id  = scaleway_lb_backend.backends[each.value.backend_name].id

  certificate_ids = length(each.value.certificate_names) > 0 ? [
    for cert_name in each.value.certificate_names : scaleway_lb_certificate.certs[cert_name].id
  ] : null

  timeout_client = each.value.timeout_client != "" ? each.value.timeout_client : null
  enable_http3   = each.value.enable_http3
}
