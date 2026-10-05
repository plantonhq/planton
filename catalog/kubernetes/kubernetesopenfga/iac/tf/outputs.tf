# Outputs — identical names and derivations in the Pulumi module
# (KubernetesOpenFgaOutputs).
#
# The service name is the chart's ClusterIP Service — openfga.fullname,
# pinned to the resource name via fullnameOverride; the endpoints are
# built from it (HTTP 8080, plaintext gRPC 8081 — the chart's fixed
# ports).

output "namespace" {
  description = "Namespace the OpenFGA server runs in"
  value       = local.namespace
}

output "service" {
  description = "Name of the OpenFGA Service (= metadata.name via fullnameOverride)"
  value       = local.service_name
}

output "api_http_endpoint" {
  description = "In-cluster HTTP API endpoint — the REST surface SDKs and the platform's OpenFGA provider connect to"
  value       = "http://${local.service_name}.${local.namespace}.svc.cluster.local:8080"
}

output "api_grpc_endpoint" {
  description = "In-cluster gRPC API endpoint host:port (plaintext gRPC)"
  value       = "${local.service_name}.${local.namespace}.svc.cluster.local:8081"
}

output "authn_keys_secret_name" {
  description = "Module-owned Secret holding declared pre-shared API keys (data key `keys`); empty when authn is unset or rides an existing Secret"
  value       = local.authn_keys_secret_name
}

output "port_forward_command" {
  description = "Copy-paste command for reaching the HTTP API from a workstation"
  value       = "kubectl port-forward -n ${local.namespace} svc/${local.service_name} 8080:8080"
}
