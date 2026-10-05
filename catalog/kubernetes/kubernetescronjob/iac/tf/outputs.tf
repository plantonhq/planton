# Outputs — must flatten onto KubernetesCronJobOutputs
# (outputs.proto) identically to the Pulumi module's exports.

output "namespace" {
  description = "The namespace the CronJob was created in"
  value       = local.namespace
}

output "cron_job_name" {
  description = "The name of the CronJob object as created in the cluster"
  value       = var.metadata.name
}

output "schedule" {
  description = "The effective cron expression the CronJob runs on — the deployed truth for dependents and audits"
  value       = var.spec.schedule
}
