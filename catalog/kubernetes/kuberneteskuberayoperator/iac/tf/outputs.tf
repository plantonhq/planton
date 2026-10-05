# Outputs — flattened onto KubernetesKubeRayOperatorOutputs by
# the platform. Keep in lockstep with the Pulumi module's exports.

output "namespace" {
  description = "Kubernetes namespace the operator was installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name of the operator (metadata.name; the chart's fullname is pinned to it, so child names derive from it)"
  value       = local.release_name
}

output "watched_namespaces" {
  description = "Namespaces the operator watches for Ray CRs (empty = cluster-wide — RayCluster declarations reconcile anywhere)"
  value       = local.watch_namespaces
}
