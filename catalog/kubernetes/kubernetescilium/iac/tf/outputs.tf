# Outputs — flattened onto KubernetesCiliumOutputs by the
# platform. Keep in lockstep with the Pulumi module's exports.

output "namespace" {
  description = "Kubernetes namespace Cilium was installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name (fixed \"cilium\" — one dataplane per cluster)"
  value       = local.release_name
}

output "cluster_name" {
  description = "Cluster identity Cilium runs under (the resolved spec.cluster_name; \"default\" when unset) — the name this cluster carries in Hubble flows and any future Cluster Mesh"
  value       = local.cluster_name
}

output "hubble_relay_service_name" {
  description = "Name of the hubble-relay Service (fixed \"hubble-relay\" by the chart) when hubble.relay is enabled; empty otherwise"
  value       = local.hubble_relay_service_name
}

output "hubble_ui_service_name" {
  description = "Name of the hubble-ui Service (fixed \"hubble-ui\" by the chart) when hubble.ui is enabled; empty otherwise"
  value       = local.hubble_ui_service_name
}

output "gateway_class_name" {
  description = "Name of the GatewayClass Cilium registers (fixed \"cilium\" by the chart) when gateway_api is enabled; empty otherwise"
  value       = local.gateway_class_name
}
