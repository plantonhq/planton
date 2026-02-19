resource "hcloud_load_balancer" "this" {
  name               = local.lb_name
  load_balancer_type = var.spec.load_balancer_type
  location           = var.spec.location
  labels             = local.standard_labels
  delete_protection  = var.spec.delete_protection != null ? var.spec.delete_protection : false

  algorithm {
    type = local.algorithm
  }
}

# --- Services (keyed by effective listen_port per CG02) ---

resource "hcloud_load_balancer_service" "this" {
  for_each = local.services

  load_balancer_id = hcloud_load_balancer.this.id
  protocol         = each.value.protocol
  listen_port      = each.value.effective_listen_port
  destination_port = each.value.effective_destination_port
  proxyprotocol    = each.value.proxyprotocol != null ? each.value.proxyprotocol : false

  dynamic "http" {
    for_each = (
      each.value.http != null && each.value.protocol != "tcp"
      ? [each.value.http]
      : []
    )
    content {
      sticky_sessions = http.value.sticky_sessions != null ? http.value.sticky_sessions : false
      cookie_name     = http.value.cookie_name
      cookie_lifetime = http.value.cookie_lifetime
      certificates = (
        http.value.certificate_ids != null
        ? [for id in http.value.certificate_ids : tonumber(id)]
        : []
      )
      redirect_http = http.value.redirect_http != null ? http.value.redirect_http : false
    }
  }

  dynamic "health_check" {
    for_each = each.value.health_check != null ? [each.value.health_check] : []
    content {
      protocol = (
        health_check.value.protocol != null && health_check.value.protocol != "" && health_check.value.protocol != "service_protocol_unspecified"
        ? health_check.value.protocol
        : (each.value.protocol == "https" ? "http" : each.value.protocol)
      )
      port = (
        health_check.value.port != null
        ? health_check.value.port
        : each.value.effective_destination_port
      )
      interval = health_check.value.interval != null ? health_check.value.interval : 15
      timeout  = health_check.value.timeout != null ? health_check.value.timeout : 10
      retries  = health_check.value.retries != null ? health_check.value.retries : 3

      dynamic "http" {
        for_each = health_check.value.http != null ? [health_check.value.http] : []
        content {
          domain       = http.value.domain
          path         = http.value.path
          response     = http.value.response
          tls          = http.value.tls
          status_codes = http.value.status_codes
        }
      }
    }
  }
}

# --- Server Targets (keyed by server_id per CG02) ---

resource "hcloud_load_balancer_target" "server" {
  for_each = {
    for t in (var.spec.server_targets != null ? var.spec.server_targets : []) :
    t.server_id => t
  }

  load_balancer_id = hcloud_load_balancer.this.id
  type             = "server"
  server_id        = tonumber(each.value.server_id)
  use_private_ip   = each.value.use_private_ip != null ? each.value.use_private_ip : false

  depends_on = [hcloud_load_balancer_network.this]
}

# --- Label Selector Targets (keyed by selector per CG02) ---

resource "hcloud_load_balancer_target" "label_selector" {
  for_each = {
    for t in (var.spec.label_selector_targets != null ? var.spec.label_selector_targets : []) :
    t.selector => t
  }

  load_balancer_id = hcloud_load_balancer.this.id
  type             = "label_selector"
  label_selector   = each.value.selector
  use_private_ip   = each.value.use_private_ip != null ? each.value.use_private_ip : false

  depends_on = [hcloud_load_balancer_network.this]
}

# --- IP Targets (keyed by ip per CG02) ---

resource "hcloud_load_balancer_target" "ip" {
  for_each = {
    for t in (var.spec.ip_targets != null ? var.spec.ip_targets : []) :
    t.ip => t
  }

  load_balancer_id = hcloud_load_balancer.this.id
  type             = "ip"
  ip               = each.value.ip
}

# --- Network Attachment (0 or 1) ---

resource "hcloud_load_balancer_network" "this" {
  count = var.spec.network != null ? 1 : 0

  load_balancer_id        = hcloud_load_balancer.this.id
  network_id              = tonumber(var.spec.network.network_id)
  ip                      = var.spec.network.ip
  enable_public_interface = var.spec.network.enable_public_interface != null ? var.spec.network.enable_public_interface : true
}
