# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports. Every child name derives from the fullname
# pinned to the resource name via fullnameOverride.

output "namespace" {
  description = "Kubernetes namespace Loki runs in"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name (= metadata.name)"
  value       = local.release_name
}

output "gateway_service" {
  description = "Name of the gateway Service (<name>-gateway, port 80); empty when the gateway is disabled"
  value       = local.gateway_enabled ? "${local.release_name}-gateway" : ""
}

output "gateway_endpoint" {
  description = "In-cluster endpoint of the gateway — the URL log shippers push to and Grafana loki datasources read from"
  value       = local.gateway_enabled ? "http://${local.release_name}-gateway.${local.namespace}.svc.cluster.local" : ""
}

output "otlp_push_endpoint" {
  description = "In-cluster OTLP log-ingest endpoint (the gateway's /otlp route)"
  value       = local.gateway_enabled ? "http://${local.release_name}-gateway.${local.namespace}.svc.cluster.local/otlp" : ""
}

output "loki_service" {
  description = "Name of the Loki HTTP Service (<name>, port 3100)"
  value       = local.release_name
}

output "port_forward_command" {
  description = "kubectl one-liner for reaching the gateway from a workstation"
  value       = local.gateway_enabled ? "kubectl port-forward svc/${local.release_name}-gateway -n ${local.namespace} 3100:80" : ""
}
