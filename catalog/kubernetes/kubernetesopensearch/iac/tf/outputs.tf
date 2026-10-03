# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports (KubernetesOpenSearchOutputs).

output "namespace" {
  description = "Namespace the cluster runs in"
  value       = local.namespace
}

output "cluster_name" {
  description = "Name of the OpenSearchCluster resource (equals metadata.name) — every operator-created object is derived from it"
  value       = local.cluster_name
}

output "service_name" {
  description = "Name of the cluster's main Service (all nodes) — the module pins general.serviceName to metadata.name"
  value       = local.cluster_name
}

output "http_endpoint" {
  description = "In-cluster HTTP API endpoint — https because the operator's HTTP layer serves TLS in every posture (operator-generated certificates, or the image's demo security configuration when spec.security is absent)"
  value       = local.http_endpoint
}

output "admin_credentials_secret_name" {
  description = "Operator-generated Secret holding the admin credentials (fields username/password) — empty when a custom security config replaces the operator bootstrap"
  value       = local.admin_credentials_secret_name
}

output "dashboards_service_name" {
  description = "Name of the Dashboards Service (`<name>-dashboards`) — empty when dashboards are not enabled"
  value       = local.dashboards_service_name
}

output "dashboards_endpoint" {
  description = "In-cluster Dashboards endpoint on port 5601 (https when dashboards TLS is enabled) — empty when dashboards are not enabled"
  value       = local.dashboards_endpoint
}

output "port_forward_command" {
  description = "Port-forward command for reaching the HTTP API from a workstation when no exposure is composed"
  value       = local.port_forward_command
}
