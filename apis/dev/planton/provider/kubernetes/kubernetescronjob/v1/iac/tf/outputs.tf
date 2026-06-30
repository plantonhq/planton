output "namespace" {
  description = "The Kubernetes namespace where the CronJob is deployed"
  value       = local.namespace
}

output "cronjob_name" {
  description = "The name of the CronJob resource"
  value       = kubernetes_cron_job_v1.this.metadata[0].name
}

output "service_account_name" {
  description = "The service account used by the CronJob"
  value       = kubernetes_service_account.this.metadata[0].name
}

output "resource_id" {
  description = "The unique resource ID for this CronJob"
  value       = local.resource_id
}

output "schedule" {
  description = "The cron schedule for the CronJob"
  value       = var.spec.schedule
}
