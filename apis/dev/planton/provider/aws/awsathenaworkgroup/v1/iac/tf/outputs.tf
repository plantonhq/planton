output "workgroup_arn" {
  description = "ARN of the Athena workgroup"
  value       = aws_athena_workgroup.this.arn
}

output "workgroup_name" {
  description = "Name of the Athena workgroup"
  value       = aws_athena_workgroup.this.name
}

output "effective_engine_version" {
  description = "Actual engine version in use by the workgroup"
  value       = try(aws_athena_workgroup.this.configuration[0].engine_version[0].effective_engine_version, "")
}
