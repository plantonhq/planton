# Outputs — flattened onto KubernetesClusterAutoscalerOutputs by
# the platform. Keep in lockstep with the Pulumi module's exports.

output "namespace" {
  description = "Kubernetes namespace the Cluster Autoscaler was installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name (fixed \"cluster-autoscaler\" — one installation per cluster; the leader-elected autoscaler owns the cluster-wide scaling decision)"
  value       = local.release_name
}

output "service_account_name" {
  # Derived from the chart's fullname template, verified in _helpers.tpl:
  # with rbac.serviceAccount.name unset the service account takes the
  # fullname, whose default name is "<cloudProvider>-<chartName>" (NOT the
  # bare chart name) — that never equals the release name, so fullname
  # renders "<release>-<cloudProvider>-<chartName>".
  description = "Name of the autoscaler's Kubernetes service account (\"cluster-autoscaler-<cloudProvider>-cluster-autoscaler\") — the subject cloud-side keyless bindings (IRSA trust policies, GCP WI bindings, Entra federated credentials) are written against"
  value       = local.service_account_name
}
