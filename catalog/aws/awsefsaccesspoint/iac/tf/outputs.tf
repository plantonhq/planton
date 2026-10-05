# ---------------------------------------------------------------------------
# Outputs -- matching AwsEfsAccessPointOutputs
# ---------------------------------------------------------------------------
# Primary consumers: Lambda (file_system_config needs the ARN), ECS task
# definitions (EFS volume authorization needs the ID).
# ---------------------------------------------------------------------------

output "access_point_id" {
  description = "The ID of the access point (e.g., fsap-0123456789abcdef0)."
  value       = aws_efs_access_point.this.id
}

output "access_point_arn" {
  description = "The ARN of the access point (Lambda file system configs and IAM policy conditions)."
  value       = aws_efs_access_point.this.arn
}

output "file_system_id" {
  description = "The ID of the file system this access point enters."
  value       = aws_efs_access_point.this.file_system_id
}

output "file_system_arn" {
  description = "The ARN of the file system this access point enters."
  value       = aws_efs_access_point.this.file_system_arn
}
