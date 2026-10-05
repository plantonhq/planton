# Outputs — must flatten onto KubernetesJobOutputs
# (outputs.proto) identically to the Pulumi module's exports.

output "namespace" {
  description = "The namespace the Job was created in"
  value       = local.namespace
}

output "job_name" {
  description = "The name of the Job object as created in the cluster"
  value       = var.metadata.name
}

output "selector_labels" {
  description = "The pod-location labels as a sorted k=v,k=v string — ready for locating this Job's pods with kubectl get pods -l or kubectl logs -l"
  value       = local.selector_labels_string
}
