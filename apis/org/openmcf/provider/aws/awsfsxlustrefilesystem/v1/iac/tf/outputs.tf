# ---------------------------------------------------------------------------
# Stack Outputs — matching AwsFsxLustreFileSystemStackOutputs
# ---------------------------------------------------------------------------
# Primary consumers: EKS (PV via FSx CSI driver), ECS (task def), EC2 (Lustre
# mount), Batch (compute environments), data repository associations.
#
# Mount command: mount -t lustre <dns_name>@tcp:/<mount_name> /mnt/fsx
# ---------------------------------------------------------------------------

output "file_system_id" {
  description = "The ID of the file system (e.g., fs-0123456789abcdef0)."
  value       = aws_fsx_lustre_file_system.this.id
}

output "file_system_arn" {
  description = "The ARN of the file system for IAM resource-level permissions."
  value       = aws_fsx_lustre_file_system.this.arn
}

output "dns_name" {
  description = "DNS name for Lustre mount (e.g., fs-xxx.fsx.region.amazonaws.com)."
  value       = aws_fsx_lustre_file_system.this.dns_name
}

output "mount_name" {
  description = "Lustre mount name (e.g., fsx or 2p5wpbwj). Used in mount command."
  value       = aws_fsx_lustre_file_system.this.mount_name
}

output "network_interface_ids" {
  description = "Network interface IDs created for the file system."
  value       = aws_fsx_lustre_file_system.this.network_interface_ids
}

output "vpc_id" {
  description = "VPC ID in which the file system was created."
  value       = aws_fsx_lustre_file_system.this.vpc_id
}

output "file_system_type_version" {
  description = "The actual Lustre version deployed (e.g., 2.12, 2.15)."
  value       = aws_fsx_lustre_file_system.this.file_system_type_version
}

output "owner_id" {
  description = "AWS account ID of the file system owner."
  value       = aws_fsx_lustre_file_system.this.owner_id
}
