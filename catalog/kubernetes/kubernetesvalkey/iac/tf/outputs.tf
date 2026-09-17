# Stack outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.
#
# Two Service handles are conditional on topology because the chart only
# renders those Services in replication mode: `<name>-read` needs
# replication with the read service enabled, and `<name>-headless` exists
# only alongside the StatefulSet (the standalone Deployment renders no
# headless Service) — both export empty otherwise.

output "namespace" {
  description = "Kubernetes namespace the Valkey instance runs in"
  value       = local.namespace
}

output "service" {
  description = "The write Service name (targets the primary in replication mode, the one instance standalone)"
  value       = local.release_name
}

output "read_service" {
  description = "The read Service name (replication mode with the read service enabled; empty otherwise)"
  value       = local.read_service_enabled ? "${local.release_name}-read" : ""
}

output "headless_service" {
  description = "The headless Service name for direct pod discovery (replication mode only; empty standalone — the chart renders no headless Service for the standalone Deployment)"
  value       = local.replication_enabled ? "${local.release_name}-headless" : ""
}

output "kube_endpoint" {
  description = "In-cluster endpoint of the write Service"
  value       = "${local.release_name}.${local.namespace}.svc.cluster.local:${local.service_port}"
}

output "port_forward_command" {
  description = "kubectl port-forward one-liner for reaching the store from a workstation"
  value       = "kubectl port-forward svc/${local.release_name} -n ${local.namespace} ${local.service_port}:${local.service_port}"
}

# The credential handles point at the module-materialized auth Secret:
# "default" is the user plain AUTH <password> maps to (the
# application-facing credential), and its Secret key is the username (the
# auth Secret's one-key-per-user layout) — the same handle whether the
# password was declared or generated here. Empty/unset when auth is off —
# no Secret exists then.
output "username" {
  description = "The ACL username applications authenticate with (empty when auth is off)"
  value       = local.auth_enabled ? "default" : ""
}

output "password_secret" {
  description = "Secret key holding the default user's password (unset when auth is off)"
  value = local.auth_enabled ? {
    name = local.auth_secret_name
    key  = "default"
  } : null
}
