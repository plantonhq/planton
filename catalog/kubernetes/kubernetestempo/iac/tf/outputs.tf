# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports. Every child name derives from the fullname
# pinned to the resource name via fullnameOverride.

output "namespace" {
  description = "Kubernetes namespace Tempo runs in"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name (= metadata.name)"
  value       = local.release_name
}

output "service" {
  description = "Name of the Tempo Service (= the release name)"
  value       = local.release_name
}

output "http_endpoint" {
  description = "In-cluster HTTP endpoint (port 3200) — the Grafana tempo datasource / TraceQL URL"
  value       = "http://${local.release_name}.${local.namespace}.svc.cluster.local:3200"
}

output "otlp_grpc_endpoint" {
  description = "In-cluster OTLP gRPC trace-ingest endpoint (port 4317)"
  value       = "${local.release_name}.${local.namespace}.svc.cluster.local:4317"
}

output "otlp_http_endpoint" {
  description = "In-cluster OTLP HTTP trace-ingest endpoint (port 4318)"
  value       = "http://${local.release_name}.${local.namespace}.svc.cluster.local:4318"
}

output "port_forward_command" {
  description = "kubectl one-liner for reaching the Tempo API from a workstation"
  value       = "kubectl port-forward svc/${local.release_name} -n ${local.namespace} 3200:3200"
}
