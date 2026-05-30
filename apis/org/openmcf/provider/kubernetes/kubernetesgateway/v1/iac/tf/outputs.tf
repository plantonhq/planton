output "gateway_name" {
  description = "Name of the created Gateway (equals metadata.name); the target of Route parentRefs"
  value       = local.gateway_name
}

output "namespace" {
  description = "Namespace the Gateway was created in"
  value       = local.namespace
}

output "gateway_class_name" {
  description = "Name of the GatewayClass this Gateway belongs to"
  value       = local.gateway_class_name
}
