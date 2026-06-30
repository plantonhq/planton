##############################################
# outputs.tf
#
# Output values for the KubernetesGatewayApiCrds
# module. These match the stack_outputs.proto
# definition.
##############################################

output "installed_version" {
  description = "Gateway API version that was installed"
  value       = local.version
}

output "installed_channel" {
  description = "Installation channel that was used (standard or experimental)"
  value       = local.channel_name
}

output "installed_manifest_url" {
  description = "Full URL of the Gateway API CRD bundle that was applied"
  value       = local.manifest_url
}
