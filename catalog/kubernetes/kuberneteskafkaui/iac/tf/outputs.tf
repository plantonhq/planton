# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.

output "namespace" {
  description = "Kubernetes namespace the console runs in"
  value       = local.namespace
}

# The chart's fullname for this release — mirrored from the chart's
# _helpers.tpl derivation, NOT overridden (see locals.service_name).
output "service_name" {
  description = "Name of the console Service"
  value       = local.service_name
}

output "endpoint" {
  description = "In-cluster console endpoint"
  value       = "http://${local.service_name}.${local.namespace}.svc.cluster.local:${local.service_port}"
}

output "port_forward_command" {
  description = "kubectl one-liner for reaching the console from a workstation without any exposure (local side pinned to 8080 — the Service port is often 80, unprivileged locally)"
  value       = "kubectl port-forward svc/${local.service_name} -n ${local.namespace} 8080:${local.service_port}"
}
