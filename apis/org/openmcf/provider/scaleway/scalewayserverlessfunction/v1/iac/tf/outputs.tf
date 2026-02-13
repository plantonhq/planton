# Stack outputs matching ScalewayServerlessFunctionStackOutputs proto.

output "function_id" {
  description = "The unique identifier of the deployed serverless function"
  value       = scaleway_function.function.id
}

output "namespace_id" {
  description = "The unique identifier of the function namespace"
  value       = scaleway_function_namespace.namespace.id
}

output "domain_name" {
  description = "The native Scaleway domain name for invoking the function"
  value       = scaleway_function.function.domain_name
}

# Complete outputs object matching stack_outputs.proto structure.
# Used by the OpenMCF platform to populate status.outputs.
output "outputs" {
  description = "Complete function outputs for integration with other resources"
  value = {
    function_id  = scaleway_function.function.id
    namespace_id = scaleway_function_namespace.namespace.id
    domain_name  = scaleway_function.function.domain_name
  }
}
