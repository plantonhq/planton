# Composition handles. All names derive from the fullnameOverride pin
# (= metadata.name). Root tokens and unseal keys are deliberately NOT
# outputs — `bao operator init` produces them at runtime and no
# deployment surface ever holds them. Twin of the Pulumi module's
# outputs.go.

output "namespace" {
  description = "Namespace the server runs in."
  value       = local.namespace
}

output "service" {
  description = "The main client Service (round-robins ALL server pods, sealed ones included — by design, so init/unseal can reach them)."
  value       = local.release_name
}

output "internal_service" {
  description = "The headless Service used for peer discovery and Raft cluster addresses."
  value       = "${local.release_name}-internal"
}

output "active_service" {
  description = "The active-leader Service — HA mode only, empty otherwise."
  value       = local.mode == "ha" ? "${local.release_name}-active" : ""
}

output "ui_service" {
  description = "The UI Service when ui_enabled, empty otherwise."
  value       = local.ui_enabled ? "${local.release_name}-ui" : ""
}

output "api_endpoint" {
  description = "In-cluster API endpoint, scheme included — what secret-consuming addons point at."
  value       = local.api_endpoint
}

output "port" {
  description = "API port (8200)."
  value       = tostring(local.api_port)
}

output "service_account_name" {
  description = "The server ServiceAccount — the identity to bind cloud IAM (auto-unseal) and Kubernetes-auth trust to. The chart derives it from the fullname, which the module pins to metadata.name."
  value       = local.release_name
}

output "port_forward_command" {
  description = "Copy-paste command for reaching the API from a workstation."
  value       = "kubectl port-forward -n ${local.namespace} svc/${local.release_name} ${local.api_port}:${local.api_port}"
}

# The backup handles: one name (`<name>-backup`) serves as the job's
# ServiceAccount, its OpenBao policy, and the CronJob, so the login recipe
# an operator copies from these outputs is one noun. Empty when `backup`
# is not declared.
output "backup_service_account_name" {
  description = "The backup job's ServiceAccount — bind the store's cloud IAM to it, and name it in the login recipe's bound_service_account_names."
  value       = local.backup_enabled ? local.backup_name : ""
}

output "backup_policy_name" {
  description = "The OpenBao policy the login recipe writes (read on sys/storage/raft/snapshot)."
  value       = local.backup_enabled ? local.backup_name : ""
}

output "backup_auth_role" {
  description = "The Kubernetes-auth role the backup job logs in with."
  value       = local.backup_enabled ? local.backup_auth_role : ""
}

output "backup_cron_job_name" {
  description = "The backup CronJob — `kubectl create job --from=cronjob/<this>` triggers a snapshot on demand."
  value       = local.backup_enabled ? local.backup_name : ""
}

output "restore_job_name" {
  description = "The restore Job (`<name>-restore-<8 hex>`) when `restore` is declared; its log is where the restore explains itself."
  value       = local.restore_job_name
}
