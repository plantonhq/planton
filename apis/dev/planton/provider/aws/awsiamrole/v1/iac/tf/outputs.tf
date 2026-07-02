output "role_arn" {
  description = "The ARN of the IAM role (what service integrations reference)."
  value       = aws_iam_role.this.arn
}

output "role_name" {
  description = "The name of the IAM role (what an AwsIamInstanceProfile's role field references)."
  value       = aws_iam_role.this.name
}

output "role_id" {
  description = "The stable unique ID AWS assigns to the role (AROA...)."
  value       = aws_iam_role.this.unique_id
}
