# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.

output "namespace" {
  description = "Namespace the Istio control plane is installed in"
  value       = local.namespace
}

output "istiod_service_name" {
  description = "Name of the istiod Service — the discovery address data-plane proxies connect to"
  value       = local.istiod_service_name
}

output "revision" {
  description = "The control-plane revision installed (\"default\" when no revision is named)"
  value       = local.revision
}

output "gateway_class_name" {
  description = "Name of the GatewayClass istiod serves — create a KubernetesGateway with this class and istiod provisions the gateway deployment"
  value       = "istio"
}

output "trust_domain" {
  description = "The mesh's trust domain — the prefix of principal strings in authorization policies"
  value       = local.trust_domain
}

output "dataplane_mode" {
  description = "The data plane mode the mesh was installed with (\"sidecar\" or \"ambient\")"
  value       = local.dataplane_mode
}
