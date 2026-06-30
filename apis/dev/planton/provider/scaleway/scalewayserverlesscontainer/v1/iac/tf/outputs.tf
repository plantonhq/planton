# Stack outputs matching ScalewayServerlessContainerStackOutputs proto.

output "container_id" {
  description = "The unique identifier of the deployed serverless container"
  value       = scaleway_container.container.id
}

output "namespace_id" {
  description = "The unique identifier of the container namespace"
  value       = scaleway_container_namespace.namespace.id
}

output "domain_name" {
  description = "The native Scaleway domain name for invoking the container"
  value       = scaleway_container.container.domain_name
}

# Complete outputs object matching stack_outputs.proto structure.
# Used by the Planton platform to populate status.outputs.
output "outputs" {
  description = "Complete container outputs for integration with other resources"
  value = {
    container_id = scaleway_container.container.id
    namespace_id = scaleway_container_namespace.namespace.id
    domain_name  = scaleway_container.container.domain_name
  }
}
