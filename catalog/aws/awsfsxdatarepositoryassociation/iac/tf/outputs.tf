# ---------------------------------------------------------------------------
# Outputs — matching AwsFsxDataRepositoryAssociationOutputs
# ---------------------------------------------------------------------------
# Primary consumers: FSx data repository tasks (association_id) and IAM
# resource-level policies (association_arn).
# ---------------------------------------------------------------------------

output "association_id" {
  description = "The AWS-assigned association ID (dra-...)."
  value       = aws_fsx_data_repository_association.this.association_id
}

output "association_arn" {
  description = "The ARN of the association for IAM resource-level permissions."
  value       = aws_fsx_data_repository_association.this.arn
}

output "file_system_id" {
  description = "The Lustre file system the association is attached to (fs-...)."
  value       = aws_fsx_data_repository_association.this.file_system_id
}
