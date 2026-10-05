# Outputs - identical names and derivations in the Pulumi module
# (KubernetesGoFeatureFlagOutputs). The Service is the chart's fullname,
# pinned to metadata.name; the endpoints are built from it and the resolved
# ports.

output "namespace" {
  description = "Namespace the relay runs in"
  value       = local.namespace
}

output "service" {
  description = "Name of the relay Service (= metadata.name via fullnameOverride)"
  value       = local.release_name
}

output "api_endpoint" {
  description = "In-cluster evaluation endpoint - the base URL for GO Feature Flag OpenFeature providers, OFREP clients and the REST API"
  value       = "http://${local.release_name}.${local.namespace}.svc.cluster.local:${local.port}"
}

output "monitoring_endpoint" {
  description = "In-cluster monitoring endpoint serving /health, /info and /metrics"
  value       = "http://${local.release_name}.${local.namespace}.svc.cluster.local:${local.monitoring_port}"
}

output "port_forward_command" {
  description = "Copy-paste command for reaching the evaluation API from a workstation"
  value       = "kubectl port-forward -n ${local.namespace} svc/${local.release_name} ${local.port}:${local.port}"
}
