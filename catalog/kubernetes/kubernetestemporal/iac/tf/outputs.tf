# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.
#
# The frontend Service is `<name>-frontend` and the Web UI Service
# `<name>-web` (fullnameOverride pins the chart's fullname to the resource
# name; the chart appends each component's name); the web handles are
# empty when the UI is disabled.

output "namespace" {
  description = "Kubernetes namespace Temporal runs in"
  value       = local.namespace
}

output "frontend_service" {
  description = "Name of the frontend Service (gRPC 7233 / HTTP 7243) — the handle exposure kinds route to"
  value       = local.frontend_service_name
}

output "frontend_endpoint" {
  description = "In-cluster frontend gRPC endpoint — what Temporal SDK workers and clients dial"
  value       = "${local.frontend_service_name}.${local.namespace}.svc.cluster.local:7233"
}

output "frontend_http_endpoint" {
  description = "In-cluster frontend HTTP API endpoint (Temporal's HTTP/JSON API)"
  value       = "http://${local.frontend_service_name}.${local.namespace}.svc.cluster.local:7243"
}

output "web_ui_service" {
  description = "Name of the Web UI Service; empty when the UI is disabled"
  value       = local.web_ui_service_name
}

output "web_ui_endpoint" {
  description = "In-cluster Web UI endpoint; empty when the UI is disabled"
  value       = local.web_ui_enabled ? "http://${local.web_ui_service_name}.${local.namespace}.svc.cluster.local:8080" : ""
}

output "port_forward_frontend_command" {
  description = "kubectl one-liner for reaching the frontend from a workstation when no exposure is composed"
  value       = "kubectl port-forward svc/${local.frontend_service_name} -n ${local.namespace} 7233:7233"
}

output "port_forward_web_ui_command" {
  description = "kubectl one-liner for reaching the Web UI from a workstation; empty when the UI is disabled"
  value       = local.web_ui_enabled ? "kubectl port-forward svc/${local.web_ui_service_name} -n ${local.namespace} 8080:8080" : ""
}
