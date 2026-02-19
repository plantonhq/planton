locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  display_name = coalesce(var.spec.display_name, var.metadata.name)

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciLoadBalancer"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  policy_map = {
    "round_robin"       = "ROUND_ROBIN"
    "least_connections" = "LEAST_CONNECTIONS"
    "ip_hash"           = "IP_HASH"
  }

  health_checker_protocol_map = {
    "http" = "HTTP"
    "tcp"  = "TCP"
  }

  listener_protocol_map = {
    "http"  = "HTTP"
    "http2" = "HTTP2"
    "tcp"   = "TCP"
    "grpc"  = "GRPC"
  }

  action_map = {
    "add_http_request_header"          = "ADD_HTTP_REQUEST_HEADER"
    "add_http_response_header"         = "ADD_HTTP_RESPONSE_HEADER"
    "extend_http_request_header_value"  = "EXTEND_HTTP_REQUEST_HEADER_VALUE"
    "extend_http_response_header_value" = "EXTEND_HTTP_RESPONSE_HEADER_VALUE"
    "remove_http_request_header"        = "REMOVE_HTTP_REQUEST_HEADER"
    "remove_http_response_header"       = "REMOVE_HTTP_RESPONSE_HEADER"
    "redirect"                          = "REDIRECT"
    "allow"                             = "ALLOW"
    "control_access_using_http_methods" = "CONTROL_ACCESS_USING_HTTP_METHODS"
    "http_header"                       = "HTTP_HEADER"
    "ip_based_max_connections"          = "IP_BASED_MAX_CONNECTIONS"
  }

  backend_sets_map = { for bs in var.spec.backend_sets : bs.name => bs }

  backends_flat = flatten([
    for bs in var.spec.backend_sets : [
      for be in bs.backends : {
        key         = "${bs.name}:${be.ip_address}:${be.port}"
        bs_name     = bs.name
        ip_address  = be.ip_address
        port        = be.port
        weight      = be.weight
        backup      = be.backup
        drain       = be.drain
        offline     = be.offline
        max_connections = be.max_connections
      }
    ]
  ])

  backends_map = { for be in local.backends_flat : be.key => be }

  listeners_map = { for ln in var.spec.listeners : ln.name => ln }

  certificates_map = { for c in var.spec.certificates : c.certificate_name => c }

  hostnames_map = { for h in var.spec.hostnames : h.name => h }

  rule_sets_map = { for rs in var.spec.rule_sets : rs.name => rs }
}
