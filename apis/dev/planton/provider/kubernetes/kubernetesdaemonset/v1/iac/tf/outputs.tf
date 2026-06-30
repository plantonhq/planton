output "namespace" {
  description = "The Kubernetes namespace the DaemonSet is deployed in."
  value       = local.namespace
}

output "daemonset_name" {
  description = "The name of the Kubernetes DaemonSet."
  value       = var.metadata.name
}

output "service_account_name" {
  description = "The name of the ServiceAccount used by the DaemonSet."
  value       = var.spec.create_service_account ? local.service_account_name : null
}

output "labels" {
  description = "The labels applied to the DaemonSet resources."
  value       = local.final_labels
}

