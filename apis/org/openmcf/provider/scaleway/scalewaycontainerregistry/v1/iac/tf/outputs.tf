# Stack outputs matching ScalewayContainerRegistryStackOutputs proto.

output "namespace_id" {
  description = "The unique identifier of the registry namespace (format: {region}/{uuid})"
  value       = scaleway_registry_namespace.registry.id
}

output "endpoint" {
  description = "The Docker endpoint URL for this registry namespace"
  value       = scaleway_registry_namespace.registry.endpoint
}

output "namespace_name" {
  description = "The name of the registry namespace as it exists in Scaleway"
  value       = scaleway_registry_namespace.registry.name
}

output "region" {
  description = "The region where the registry namespace is deployed"
  value       = local.region
}

# Complete outputs object matching stack_outputs.proto structure.
# Used by the OpenMCF platform to populate status.outputs.
output "outputs" {
  description = "Complete registry namespace outputs for integration with other resources"
  value = {
    namespace_id   = scaleway_registry_namespace.registry.id
    endpoint       = scaleway_registry_namespace.registry.endpoint
    namespace_name = scaleway_registry_namespace.registry.name
    region         = local.region
  }
}
