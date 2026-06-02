output "gateway_name" {
  description = "Name of the created Gateway (equals metadata.name)."
  value       = var.metadata.name
}

output "namespace" {
  description = "Namespace of the created Gateway."
  value       = var.spec.namespace
}

output "gateway_class_name" {
  description = "Gateway class name of the created Gateway."
  value       = var.spec.gatewayClassName
}
