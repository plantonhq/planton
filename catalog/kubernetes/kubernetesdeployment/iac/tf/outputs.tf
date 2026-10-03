# Outputs — must flatten onto KubernetesDeploymentOutputs
# (outputs.proto) identically to the Pulumi module's exports.

output "namespace" {
  description = "The namespace the workload was deployed into"
  value       = local.namespace
}

output "deployment_name" {
  description = "The name of the Deployment object as created in the cluster"
  value       = var.metadata.name
}

output "service" {
  description = "The Kubernetes Service fronting the replicas; empty when the app container exposes no ports"
  value       = local.create_service ? local.kube_service_name : ""
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
