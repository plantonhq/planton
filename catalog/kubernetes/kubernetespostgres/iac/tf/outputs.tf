# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports (KubernetesPostgresOutputs).

output "namespace" {
  description = "Namespace the cluster runs in"
  value       = local.namespace
}

output "cluster_name" {
  description = "Name of the Cluster resource (equals metadata.name) — every derived object (pods, services, secrets) is prefixed with it"
  value       = local.cluster_name
}

output "rw_service" {
  description = "Name of the read-write Service (`<name>-rw`) — always points at the current primary"
  value       = local.rw_service_name
}

output "ro_service" {
  description = "Name of the read-only Service (`<name>-ro`) — replicas only"
  value       = local.ro_service_name
}

output "r_service" {
  description = "Name of the any-instance read Service (`<name>-r`)"
  value       = local.r_service_name
}

output "kube_endpoint" {
  description = "In-cluster endpoint of the read-write Service — the connection host for applications in the same cluster"
  value       = local.kube_endpoint
}

output "port_forward_command" {
  description = "Port-forward command for reaching the primary from a workstation when no exposure is composed"
  value       = "kubectl port-forward svc/${local.rw_service_name} -n ${local.namespace} 5432:5432"
}

# The username/password handles point at the EFFECTIVE application secret:
# the operator-generated `<name>-app` normally, or the module-provided
# `<name>-app-provided` when initdb declared an owner password (the
# operator adopts a provided bootstrap secret instead of generating one).
output "username_secret" {
  description = "Secret key holding the application user's name"
  value = {
    name = local.effective_app_secret_name
    key  = "username"
  }
}

output "password_secret" {
  description = "Secret key holding the application user's password"
  value = {
    name = local.effective_app_secret_name
    key  = "password"
  }
}

output "backup_server_name" {
  description = "The backup series this install archives into (the folder beneath the backup destination path holding its base backups and WAL) — a recovery names it as source_server_name; empty when no backup is declared"
  value       = local.backup_server_name
}

output "superuser_secret_name" {
  description = "Name of the superuser credential Secret — populated only when superuser access is enabled (the provided secret when a password was declared, the operator's `<name>-superuser` otherwise), empty when disabled"
  value       = local.superuser_secret_name_output
}
