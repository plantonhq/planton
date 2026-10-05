# ---------------------------------------------------------------------------
# Outputs -- matching AwsElasticFileSystemOutputs
# ---------------------------------------------------------------------------
# Primary consumers: EKS (PersistentVolume), ECS (task def volumes),
# AwsEfsAccessPoint (file_system_id), EC2 (direct NFS mount).
# ---------------------------------------------------------------------------

output "file_system_id" {
  description = "The ID of the file system (e.g., fs-0123456789abcdef0)."
  value       = aws_efs_file_system.this.id
}

output "file_system_arn" {
  description = "The ARN of the file system for IAM resource-level permissions."
  value       = aws_efs_file_system.this.arn
}

output "dns_name" {
  description = "Regional DNS name for NFS mount (e.g., fs-xxx.efs.region.amazonaws.com)."
  value       = aws_efs_file_system.this.dns_name
}

output "mount_target_ids" {
  description = "Map of subnet ID to mount target ID."
  value       = { for k, v in aws_efs_mount_target.this : k => v.id }
}

output "mount_target_ips" {
  description = "Map of subnet ID to mount target IPv4 address (empty for IPV6_ONLY targets)."
  value       = { for k, v in aws_efs_mount_target.this : k => v.ip_address }
}

output "mount_target_ipv6_addresses" {
  description = "Map of subnet ID to mount target IPv6 address (populated for IPV6_ONLY / DUAL_STACK targets)."
  value       = { for k, v in aws_efs_mount_target.this : k => v.ipv6_address }
}

output "mount_target_dns_names" {
  description = "Map of subnet ID to per-AZ mount target DNS name."
  value       = { for k, v in aws_efs_mount_target.this : k => v.mount_target_dns_name }
}

output "replication_destination_file_system_id" {
  description = "File system ID of the replication destination; empty when replication is not configured."
  value       = try(aws_efs_replication_configuration.this[0].destination[0].file_system_id, "")
}
