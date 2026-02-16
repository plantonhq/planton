output "domain_id" {
  description = "ID of the SageMaker domain"
  value       = aws_sagemaker_domain.this.id
}

output "domain_arn" {
  description = "ARN of the SageMaker domain"
  value       = aws_sagemaker_domain.this.arn
}

output "domain_url" {
  description = "URL of the SageMaker domain"
  value       = aws_sagemaker_domain.this.url
}

output "home_efs_file_system_id" {
  description = "EFS file system ID created by the SageMaker domain"
  value       = aws_sagemaker_domain.this.home_efs_file_system_id
}

output "security_group_id_for_domain_boundary" {
  description = "Security group ID for the domain boundary"
  value       = aws_sagemaker_domain.this.security_group_id_for_domain_boundary
}

output "single_sign_on_application_arn" {
  description = "ARN of the SSO application created by the domain"
  value       = aws_sagemaker_domain.this.single_sign_on_application_arn
}
