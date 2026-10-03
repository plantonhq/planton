# Terraform outputs for Kubernetes ServiceAccount
# Keys mirror KubernetesServiceAccountOutputs field names.

output "service_account_name" {
  description = "The name of the created ServiceAccount — the value workloads set in spec.serviceAccountName"
  value       = kubernetes_service_account_v1.service_account.metadata[0].name
}

output "namespace" {
  description = "The namespace in which the ServiceAccount was created"
  value       = kubernetes_service_account_v1.service_account.metadata[0].namespace
}

output "rbac_subject" {
  description = "The fully-qualified RBAC subject string: system:serviceaccount:<namespace>:<name>"
  value       = local.rbac_subject
}

output "workload_identity_handle" {
  description = "The bound cloud identity handle (GCP email, IAM role ARN, or Azure client ID); empty when workload identity is not configured"
  value       = local.workload_identity_handle
}
