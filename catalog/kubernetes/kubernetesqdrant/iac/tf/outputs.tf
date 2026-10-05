# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.
#
# The service name is the chart's main ClusterIP Service — qdrant.fullname,
# pinned to the resource name via fullnameOverride. The API-key Secret names
# point at the chart-owned `<name>-apikey` Secret (keys api-key /
# read-only-api-key), which the chart populates for the generate and
# existing-secret arms alike.

output "namespace" {
  description = "Kubernetes namespace the Qdrant cluster runs in"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name (= metadata.name)"
  value       = local.release_name
}

output "service_name" {
  description = "Name of the main Qdrant Service (http 6333, grpc 6334; = the release name)"
  value       = local.service_name
}

output "http_endpoint" {
  description = "In-cluster REST endpoint (scheme follows the tls block)"
  value       = "${local.http_scheme}://${local.service_name}.${local.namespace}.svc.cluster.local:6333"
}

output "grpc_endpoint" {
  description = "In-cluster gRPC endpoint SDKs default to"
  value       = "${local.service_name}.${local.namespace}.svc.cluster.local:6334"
}

output "api_key_secret_name" {
  description = "Chart-owned Secret holding the read-write API key (key api-key); empty when unauthenticated"
  value       = local.api_key_secret_name
}

output "read_only_api_key_secret_name" {
  description = "Secret holding the read-only API key (key read-only-api-key); empty when not declared"
  value       = local.read_only_api_key_secret_name
}

output "port_forward_command" {
  description = "kubectl one-liner for reaching REST from a workstation"
  value       = "kubectl port-forward svc/${local.service_name} -n ${local.namespace} 6333:6333"
}
