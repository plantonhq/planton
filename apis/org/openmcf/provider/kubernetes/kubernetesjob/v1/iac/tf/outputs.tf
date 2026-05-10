output "namespace" {
  description = "The Kubernetes namespace where the Job is deployed"
  value       = local.namespace
}

output "job_name" {
  description = "The name of the Job resource"
  value       = kubernetes_job_v1.this.metadata[0].name
}

output "service_account_name" {
  description = "The service account used by the Job"
  value       = kubernetes_service_account.this.metadata[0].name
}

output "resource_id" {
  description = "The unique resource ID for this Job"
  value       = local.resource_id
}
