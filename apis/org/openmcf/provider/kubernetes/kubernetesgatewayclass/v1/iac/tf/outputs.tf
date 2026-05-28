output "gateway_class_name" {
  description = "Name of the created GatewayClass (equals metadata.name)"
  value       = local.gateway_class_name
}

output "controller_name" {
  description = "The controller managing this GatewayClass"
  value       = local.controller_name
}
