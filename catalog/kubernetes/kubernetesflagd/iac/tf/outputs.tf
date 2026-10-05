# Outputs - identical names and derivations in the Pulumi module
# (KubernetesFlagdOutputs).

output "namespace" {
  description = "Namespace flagd runs in"
  value       = local.namespace
}

output "service" {
  description = "Name of the flagd Service (= metadata.name)"
  value       = local.name
}

output "evaluation_endpoint" {
  description = "In-cluster gRPC evaluation endpoint host:port"
  value       = "${local.name}.${local.namespace}.svc.cluster.local:${local.port}"
}

output "sync_endpoint" {
  description = "In-cluster sync endpoint host:port for in-process providers"
  value       = "${local.name}.${local.namespace}.svc.cluster.local:${local.sync_port}"
}

output "ofrep_endpoint" {
  description = "In-cluster OFREP base URL"
  value       = "http://${local.name}.${local.namespace}.svc.cluster.local:${local.ofrep_port}"
}

output "management_endpoint" {
  description = "In-cluster management endpoint serving /healthz, /readyz and /metrics"
  value       = "http://${local.name}.${local.namespace}.svc.cluster.local:${local.management_port}"
}

output "port_forward_command" {
  description = "Copy-paste command for reaching the OFREP endpoint from a workstation"
  value       = "kubectl port-forward -n ${local.namespace} svc/${local.name} ${local.ofrep_port}:${local.ofrep_port}"
}
