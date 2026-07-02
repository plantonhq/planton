output "user_arn" {
  description = "The ARN of the IAM user."
  value       = aws_iam_user.this.arn
}

output "user_name" {
  description = "The IAM user name."
  value       = aws_iam_user.this.name
}

output "user_id" {
  description = "The stable unique ID AWS assigns to the user (AIDA...)."
  value       = aws_iam_user.this.unique_id
}

output "access_key_id" {
  description = "Access key ID (if created)."
  value       = try(aws_iam_access_key.this[0].id, "")
  sensitive   = true
}

output "secret_access_key" {
  # Base64-encoded to match the stack-outputs contract (the proto documents the
  # secret as base64), keeping both engines' outputs byte-identical.
  description = "Base64-encoded secret access key (if created)."
  value       = try(base64encode(aws_iam_access_key.this[0].secret), "")
  sensitive   = true
}

output "console_url" {
  description = "AWS console sign-in URL for this user."
  value       = "https://signin.aws.amazon.com/console"
}
