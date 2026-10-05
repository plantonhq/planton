# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.
#
# The server service is `<name>-server` (fullnameOverride pins the chart's
# fullname to the resource name; the chart appends each component's name);
# both server handles are empty when the server is disabled.

output "namespace" {
  description = "Kubernetes namespace Argo Workflows runs in"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name (= metadata.name)"
  value       = local.release_name
}

output "server_service" {
  description = "Name of the Argo server Service (port 2746); empty when the server is disabled"
  value       = local.server_service_name
}

output "server_kube_endpoint" {
  description = "In-cluster endpoint of the Argo server (scheme follows spec.server.secure); empty when the server is disabled"
  value       = local.server_enabled ? "${local.server_scheme}://${local.server_service_name}.${local.namespace}.svc.cluster.local:2746" : ""
}

output "workflow_service_account" {
  description = "ServiceAccount workflow pods run as — annotate THIS for IRSA/workload identity"
  value       = local.workflow_service_account
}

output "port_forward_command" {
  description = "kubectl one-liner for reaching the UI from a workstation; empty when the server is disabled"
  value       = local.server_enabled ? "kubectl port-forward svc/${local.server_service_name} -n ${local.namespace} 2746:2746" : ""
}
