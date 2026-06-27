output "repository_name" {
  description = "The ECR repository name."
  value       = aws_ecr_repository.this.name
}

output "repository_url" {
  description = "The ECR repository URL."
  value       = aws_ecr_repository.this.repository_url
}

output "repository_arn" {
  description = "The ECR repository ARN."
  value       = aws_ecr_repository.this.arn
}

output "registry_id" {
  description = "The AWS account (registry) ID for the repository."
  value       = aws_ecr_repository.this.registry_id
}

output "repository_uri" {
  description = "The URI of the repository."
  value       = aws_ecr_repository.this.repository_url
}

output "lifecycle_policy" {
  description = "The lifecycle policy attached to the repository, or null when none is configured."
  value       = length(aws_ecr_lifecycle_policy.this) > 0 ? aws_ecr_lifecycle_policy.this[0].policy : null
}


