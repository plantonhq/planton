# Outputs — flattened onto KubernetesPlantonOperatorOutputs by
# the platform. Keep in lockstep with the Pulumi module's exports.

output "namespace" {
  description = "Kubernetes namespace the operator was installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name of the operator (fixed \"planton-operator\" — one installation per cluster, enforced by the operator's own startup guard)"
  value       = local.release_name
}
