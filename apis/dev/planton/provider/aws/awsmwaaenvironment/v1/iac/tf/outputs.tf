output "environment_arn" {
  description = "ARN of the MWAA environment"
  value       = aws_mwaa_environment.this.arn
}

output "environment_name" {
  description = "Name of the MWAA environment"
  value       = aws_mwaa_environment.this.name
}

output "webserver_url" {
  description = "URL of the Airflow web UI"
  value       = aws_mwaa_environment.this.webserver_url
}

output "airflow_version" {
  description = "Effective Apache Airflow version"
  value       = aws_mwaa_environment.this.airflow_version
}

output "service_role_arn" {
  description = "ARN of the AWS service role for the environment"
  value       = aws_mwaa_environment.this.service_role_arn
}

output "environment_class" {
  description = "Effective environment class"
  value       = aws_mwaa_environment.this.environment_class
}

output "status" {
  description = "Current status of the MWAA environment"
  value       = aws_mwaa_environment.this.status
}

output "security_group_id" {
  description = "ID of the managed security group (if created)"
  value       = length(aws_security_group.environment) > 0 ? aws_security_group.environment[0].id : ""
}
