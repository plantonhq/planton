# Outputs — flattened onto KubernetesKarpenterOutputs by the
# platform. Keep in lockstep with the Pulumi module's exports.

output "namespace" {
  description = "Kubernetes namespace Karpenter was installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Controller Helm release name (fixed \"karpenter\" — one installation per cluster; Karpenter owns the cluster-wide karpenter.sh label domain and node lifecycle)"
  value       = local.release_name
}

output "crd_release_name" {
  description = "CRD Helm release name (fixed \"karpenter-crd\"; empty when spec.crds.install is false and something else manages the CRDs)"
  value       = local.crds_install ? local.crd_release_name : ""
}

output "service_account_name" {
  description = "Name of the controller's Kubernetes service account (fixed \"karpenter\" — the chart's serviceAccount.name defaults to the fullname template, which resolves to the release name because the release name contains the chart name); the subject IRSA trust policies and EKS Pod Identity associations are written against"
  value       = "karpenter"
}
