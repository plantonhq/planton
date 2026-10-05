# Outputs — flattened onto KubernetesKedaOutputs by the
# platform. Keep in lockstep with the Pulumi module's exports.

output "namespace" {
  description = "Kubernetes namespace KEDA was installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name (fixed \"keda\" — one installation per cluster; the external.metrics.k8s.io APIService is a cluster singleton)"
  value       = local.release_name
}

output "operator_service_account_name" {
  description = "Name of the operator's Kubernetes service account (the chart's fixed \"keda-operator\") — the subject cloud-side keyless bindings are written against"
  value       = "keda-operator"
}
