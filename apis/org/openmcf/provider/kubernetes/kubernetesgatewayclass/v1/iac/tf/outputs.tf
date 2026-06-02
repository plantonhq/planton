output "gateway_class_name" {
  description = "Name of the created GatewayClass (equals metadata.name)."
  value       = var.metadata.name
}

output "controller_name" {
  description = "Controller name of the created GatewayClass."
  value       = var.spec.controllerName
}
