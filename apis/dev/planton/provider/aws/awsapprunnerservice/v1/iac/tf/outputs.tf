output "service_arn" {
  description = "Full ARN of the App Runner Service"
  value       = aws_apprunner_service.this.arn
}

output "service_id" {
  description = "Unique identifier assigned by App Runner"
  value       = aws_apprunner_service.this.service_id
}

output "service_url" {
  description = "Public HTTPS URL for the App Runner Service"
  value       = aws_apprunner_service.this.service_url
}

output "service_name" {
  description = "Computed name of the App Runner Service"
  value       = aws_apprunner_service.this.service_name
}

output "service_status" {
  description = "Current operational status of the service"
  value       = aws_apprunner_service.this.status
}

output "vpc_connector_arn" {
  description = "ARN of the VPC Connector (empty when default egress is used)"
  value = (
    local.create_inline_vpc_connector
    ? aws_apprunner_vpc_connector.this[0].arn
    : (local.use_external_vpc_connector ? var.spec.vpc_connector_arn : "")
  )
}

output "auto_scaling_configuration_arn" {
  description = "ARN of the Auto Scaling Configuration Version (empty when not configured)"
  value = (
    var.spec.auto_scaling != null
    ? aws_apprunner_auto_scaling_configuration_version.this[0].arn
    : ""
  )
}
