# ---------------------------------------------------------------------------
# Stack Outputs — matching AwsElasticFileSystemStackOutputs
# ---------------------------------------------------------------------------
# Primary consumers: EKS (PersistentVolume), ECS (task def), Lambda, EC2 (NFS).
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
  description = "Map of subnet ID to mount target IP address."
  value       = { for k, v in aws_efs_mount_target.this : k => v.ip_address }
}

output "mount_target_dns_names" {
  description = "Map of subnet ID to per-AZ mount target DNS name."
  value       = { for k, v in aws_efs_mount_target.this : k => v.mount_target_dns_name }
}

output "access_point_ids" {
  description = "Map of access point name to access point ID."
  value       = { for k, v in aws_efs_access_point.this : k => v.id }
}

output "access_point_arns" {
  description = "Map of access point name to access point ARN."
  value       = { for k, v in aws_efs_access_point.this : k => v.arn }
}
