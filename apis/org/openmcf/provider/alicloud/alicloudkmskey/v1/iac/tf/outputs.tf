output "key_id" {
  description = "The KMS key ID"
  value       = alicloud_kms_key.main.id
}

output "arn" {
  description = "The KMS key ARN"
  value       = alicloud_kms_key.main.arn
}
