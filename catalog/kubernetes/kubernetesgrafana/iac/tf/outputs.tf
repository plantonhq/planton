# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.
#
# The service name is the chart's ClusterIP Service — grafana.fullname,
# pinned to the resource name via fullnameOverride. The admin Secret name
# follows the credential arm: the chart-owned `<name>` Secret for the
# generate arm, the referenced Secret's own name for the existing arm.

output "namespace" {
  description = "Kubernetes namespace Grafana runs in"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name (= metadata.name)"
  value       = local.release_name
}

output "service" {
  description = "Name of the Grafana Service (port 80 → container 3000; = the release name)"
  value       = local.service_name
}

output "endpoint" {
  description = "In-cluster endpoint (plain HTTP on the Service; TLS composes at the exposure layer)"
  value       = "http://${local.service_name}.${local.namespace}.svc.cluster.local"
}

output "admin_secret_name" {
  description = "Secret holding the admin credentials (keys admin-user / admin-password for the chart-generated arm)"
  value       = local.admin_secret_name
}

output "port_forward_command" {
  description = "kubectl one-liner for reaching the UI from a workstation"
  value       = "kubectl port-forward svc/${local.service_name} -n ${local.namespace} 3000:80"
}
