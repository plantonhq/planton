output "instance_profile_arn" {
  description = "The ARN of the instance profile (what an EC2 instance's iam_instance_profile_arn references)."
  value       = aws_iam_instance_profile.this.arn
}

output "instance_profile_name" {
  description = "The friendly name of the instance profile (launch templates take the profile by name)."
  value       = aws_iam_instance_profile.this.name
}

output "instance_profile_id" {
  description = "The stable unique ID AWS assigns to the profile (AIPA...)."
  value       = aws_iam_instance_profile.this.unique_id
}

output "role_name" {
  description = "The name of the IAM role the profile carries."
  value       = aws_iam_instance_profile.this.role
}
