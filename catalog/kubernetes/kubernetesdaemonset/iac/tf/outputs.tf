# Outputs — must flatten onto KubernetesDaemonSetOutputs
# (outputs.proto) identically to the Pulumi module's exports.
# DaemonSets have no Service or ingress, so the composition surface is the
# object identity and its selector labels.

output "namespace" {
  description = "The namespace the workload was deployed into"
  value       = local.namespace
}

output "daemon_set_name" {
  description = "The name of the DaemonSet object as created in the cluster"
  value       = var.metadata.name
}

output "selector_labels" {
  description = "The pod selector labels as a sorted k=v,k=v string — ready for NetworkPolicy podSelectors, kubectl -l, and pod-affinity terms"
  value       = local.selector_labels_string
}
