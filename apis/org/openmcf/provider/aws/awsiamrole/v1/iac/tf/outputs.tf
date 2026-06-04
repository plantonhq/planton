output "role_arn" {
  description = "The ARN of the IAM role."
  value       = aws_iam_role.this.arn
}

output "role_name" {
  description = "The name of the IAM role."
  value       = aws_iam_role.this.name
}

output "instance_profile_arn" {
  description = "The ARN of the IAM instance profile wrapping this role (attach to EC2 instances)."
  value       = aws_iam_instance_profile.this.arn
}

output "instance_profile_name" {
  description = "The name of the IAM instance profile wrapping this role."
  value       = aws_iam_instance_profile.this.name
}



