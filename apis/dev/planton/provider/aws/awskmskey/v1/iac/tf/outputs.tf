output "key_id" {
  description = "The KMS key ID (UUID)."
  value       = aws_kms_key.this.key_id
}

output "key_arn" {
  description = "The KMS key ARN."
  value       = aws_kms_key.this.arn
}

output "alias_name" {
  description = "Alias name assigned to the KMS key, if any."
  value       = local.alias_name != null && local.alias_name != "" ? local.alias_name : ""
}

output "rotation_enabled" {
  description = "Whether key rotation is enabled."
  value       = local.rotation_enabled
}



