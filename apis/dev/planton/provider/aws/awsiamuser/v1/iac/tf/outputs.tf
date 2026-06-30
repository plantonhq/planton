output "user_arn" {
  description = "The ARN of the IAM user."
  value       = aws_iam_user.this.arn
}

output "access_key_id" {
  description = "Access key ID (if created)."
  value       = try(aws_iam_access_key.this[0].id, "")
  sensitive   = true
}

output "secret_access_key" {
  description = "Secret access key (if created)."
  value       = try(aws_iam_access_key.this[0].secret, "")
  sensitive   = true
}

output "console_url" {
  description = "AWS console sign-in URL for this user."
  value       = ""
}

output "user_name" {
  description = "The IAM user name."
  value       = aws_iam_user.this.name
}

output "user_id" {
  description = "The IAM user unique ID."
  value       = aws_iam_user.this.unique_id
}



