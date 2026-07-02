output "policy_arn" {
  description = "The ARN of the managed policy (what attachments and permissions boundaries reference)."
  value       = aws_iam_policy.this.arn
}

output "policy_id" {
  description = "The stable unique ID AWS assigns to the policy (ANPA...)."
  value       = aws_iam_policy.this.policy_id
}

output "policy_name" {
  description = "The friendly name of the policy."
  value       = aws_iam_policy.this.name
}
