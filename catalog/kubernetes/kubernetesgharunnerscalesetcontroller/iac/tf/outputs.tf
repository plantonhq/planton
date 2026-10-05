# Outputs — identical names and derivations in the Pulumi module
# (KubernetesGhaRunnerScaleSetControllerOutputs).

output "namespace" {
  description = "Namespace the controller is installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name (equals metadata.name)"
  value       = local.release_name
}

output "service_account_name" {
  description = "Name of the controller's ServiceAccount — what a KubernetesGhaRunnerScaleSet references in controller_service_account when this controller watches a single namespace"
  value       = local.service_account_name
}
