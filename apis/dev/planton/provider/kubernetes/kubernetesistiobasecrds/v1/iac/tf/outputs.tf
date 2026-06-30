##############################################
# outputs.tf
#
# Output values for the KubernetesIstioBaseCrds
# module. These match the stack_outputs.proto
# definition.
##############################################

output "installed_release" {
  description = "Istio release the CRDs were installed from"
  value       = local.istio_release
}

output "installed_manifest_url" {
  description = "Full URL of the istio/base CRD bundle that was applied"
  value       = local.manifest_url
}
