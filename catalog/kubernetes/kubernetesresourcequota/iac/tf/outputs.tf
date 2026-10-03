# Outputs — must flatten onto KubernetesResourceQuotaOutputs
# (outputs.proto) identically to the Pulumi module's exports.

output "resource_quota_name" {
  description = "The name of the ResourceQuota object as created in the cluster"
  value       = kubernetes_resource_quota_v1.resource_quota.metadata[0].name
}

output "namespace" {
  description = "The namespace the quota governs"
  value       = kubernetes_resource_quota_v1.resource_quota.metadata[0].namespace
}

output "limit_range_name" {
  description = "The name of the companion LimitRange; empty when no limit_defaults were configured"
  value       = local.limit_range_name
}
