# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.
#
# The server service is `<name>-server` (fullnameOverride pins the chart's
# fullname to the resource name; the chart appends each component's name).
# The initial admin Secret's name is fixed by the APPLICATION — Argo CD
# creates `argocd-initial-admin-secret` in its namespace at first start —
# and is exported only while the admin user is enabled.

output "namespace" {
  description = "Kubernetes namespace Argo CD runs in"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name (= metadata.name)"
  value       = local.release_name
}

output "server_service" {
  description = "Name of the Argo CD API/UI server Service (port 80 http / 443 https)"
  value       = local.server_service_name
}

output "server_kube_endpoint" {
  description = "In-cluster endpoint of the server (scheme follows spec.server.insecure)"
  value       = "${local.server_scheme}://${local.server_service_name}.${local.namespace}.svc.cluster.local"
}

output "initial_admin_secret_name" {
  description = "Secret holding the generated initial admin password (key `password`); empty when the admin user is disabled"
  value       = local.initial_admin_secret_name
}

output "port_forward_command" {
  description = "kubectl one-liner for reaching the UI from a workstation"
  value       = "kubectl port-forward svc/${local.server_service_name} -n ${local.namespace} 8080:${local.server_insecure ? 80 : 443}"
}
