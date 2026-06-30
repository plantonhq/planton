# Terraform outputs for Kubernetes Namespace

output "namespace" {
  description = "The created namespace name"
  value       = kubernetes_namespace_v1.namespace.metadata[0].name
}

output "namespace_id" {
  description = "Namespace identifier"
  value       = kubernetes_namespace_v1.namespace.metadata[0].name
}

output "resource_quotas_applied" {
  description = "Whether resource quotas were configured"
  value       = local.resource_quota_enabled
}

output "limit_ranges_applied" {
  description = "Whether default limits were set"
  value       = local.limit_range_enabled
}

output "network_policies_applied" {
  description = "Whether network policies were created"
  value       = local.isolate_ingress || local.restrict_egress
}

output "service_mesh_enabled" {
  description = "Service mesh injection status"
  value       = local.service_mesh_enabled
}

output "service_mesh_type" {
  description = "The configured mesh type"
  value       = local.service_mesh_type
}

output "pod_security_standard" {
  description = "Enforced security level"
  value       = local.pod_security_standard
}

output "labels_json" {
  description = "Applied labels (JSON)"
  value       = jsonencode(local.labels)
}

output "annotations_json" {
  description = "Applied annotations (JSON)"
  value       = jsonencode(local.annotations)
}


