# Outputs — must flatten onto KubernetesStatefulSetOutputs
# (outputs.proto) identically to the Pulumi module's exports.

output "namespace" {
  description = "The namespace the workload was deployed into"
  value       = local.namespace
}

output "stateful_set_name" {
  description = "The name of the StatefulSet object as created in the cluster"
  value       = var.metadata.name
}

output "service" {
  description = "The headless governing Service of the StatefulSet — the Service that gives each replica its stable per-pod DNS name; load-balanced client access also goes through this name"
  value       = local.kube_service_name
}

output "selector_labels" {
  description = "The pod selector labels as a sorted k=v,k=v string — ready for NetworkPolicy podSelectors, kubectl -l, and pod-affinity terms"
  value       = local.selector_labels_string
}

output "port_forward_command" {
  description = "Ready-to-run port-forward command for reaching the workload without external exposure"
  value       = local.kube_port_forward_command
}

output "kube_endpoint" {
  description = "In-cluster DNS endpoint of the Service — the handle exposure kinds and sibling workloads connect to"
  value       = local.kube_service_fqdn
}

output "pod_dns_template" {
  description = "Template for each replica's stable DNS name; {ordinal} is a literal placeholder the consumer substitutes with the replica index (e.g. \"0\") to address a specific member"
  value       = local.pod_dns_template
}
