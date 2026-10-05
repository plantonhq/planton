# Outputs — flattened onto KubernetesStrimziKafkaOperatorOutputs
# by the platform. Keep in lockstep with the Pulumi module's exports.

output "namespace" {
  description = "Kubernetes namespace the operator was installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name of the operator (metadata.name)"
  value       = local.release_name
}
