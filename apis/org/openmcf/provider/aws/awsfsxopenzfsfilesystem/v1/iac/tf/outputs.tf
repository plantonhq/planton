# ---------------------------------------------------------------------------
# Stack Outputs — matching AwsFsxOpenzfsFileSystemStackOutputs
# ---------------------------------------------------------------------------
# Primary consumers: EKS (PV via FSx OpenZFS CSI driver), ECS (task def NFS
# volumes), EC2 (direct NFS mount), child volume creation (root_volume_id).
#
# Mount command: mount -t nfs <dns_name>:/fsx /mnt/fsx
# ---------------------------------------------------------------------------

output "file_system_id" {
  description = "The ID of the file system (e.g., fs-0123456789abcdef0)."
  value       = aws_fsx_openzfs_file_system.this.id
}

output "file_system_arn" {
  description = "The ARN of the file system for IAM resource-level permissions."
  value       = aws_fsx_openzfs_file_system.this.arn
}

output "dns_name" {
  description = "DNS name for NFS mount (e.g., fs-xxx.fsx.region.amazonaws.com)."
  value       = aws_fsx_openzfs_file_system.this.dns_name
}

output "endpoint_ip_address" {
  description = "Endpoint IP address. Floating IP for MULTI_AZ failover."
  value       = aws_fsx_openzfs_file_system.this.endpoint_ip_address
}

output "root_volume_id" {
  description = "Root volume ID. Use as parent_volume_id for child volumes."
  value       = aws_fsx_openzfs_file_system.this.root_volume_id
}

output "network_interface_ids" {
  description = "Network interface IDs created for the file system."
  value       = aws_fsx_openzfs_file_system.this.network_interface_ids
}

output "vpc_id" {
  description = "VPC ID in which the file system was created."
  value       = aws_fsx_openzfs_file_system.this.vpc_id
}

output "owner_id" {
  description = "AWS account ID of the file system owner."
  value       = aws_fsx_openzfs_file_system.this.owner_id
}
