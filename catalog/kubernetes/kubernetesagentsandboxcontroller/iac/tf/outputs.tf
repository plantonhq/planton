# Outputs, identical in name and derivation to the Pulumi module's
# (outputs.go outputValues).

locals {
  router_host = "${local.router_service_name}.${local.namespace}.svc.${local.cluster_domain}"
}

output "namespace" {
  description = "Namespace the controller and router run in (always agent-sandbox-system)"
  value       = local.namespace
}

output "metrics_service_name" {
  description = "Service in front of the controller's metrics endpoint"
  value       = local.controller_service_name
}

output "metrics_port" {
  description = "Port of the controller's metrics Service"
  value       = local.c_metrics_port
}

output "router_service_name" {
  description = "The router's Service (sandbox-router-svc); empty without a router"
  value       = local.router_on ? local.router_service_name : ""
}

output "router_endpoint" {
  description = "URL a product or SDK dials to reach a sandbox through the router; empty without a router"
  value = !local.router_on ? "" : (
    local.r_http_enabled ? "http://${local.router_host}:${local.r_http_port}" : "https://${local.router_host}:${local.r_tls_port}"
  )
}
