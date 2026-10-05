# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.

output "namespace" {
  description = "Kubernetes namespace Velero was installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name (fixed \"velero\" — one installation per cluster)"
  value       = helm_release.velero.name
}

output "service_account_name" {
  description = "Chart-derived name of the Velero server ServiceAccount (\"velero-server\") — the subject cloud-side keyless bindings (IRSA trust policies, GCP WI bindings, Azure federated credentials) are written against"
  value       = local.service_account_name
}

output "backup_storage_location_name" {
  description = "Name of the default BackupStorageLocation (\"default\") — what Backup and Schedule resources reference through storageLocation"
  value       = local.backup_storage_location_name
}
