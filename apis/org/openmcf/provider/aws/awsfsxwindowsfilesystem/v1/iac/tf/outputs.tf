output "file_system_id" {
  description = "The ID of the file system"
  value       = aws_fsx_windows_file_system.this.id
}

output "file_system_arn" {
  description = "The Amazon Resource Name of the file system"
  value       = aws_fsx_windows_file_system.this.arn
}

output "dns_name" {
  description = "The DNS name for the file system"
  value       = aws_fsx_windows_file_system.this.dns_name
}

output "preferred_file_server_ip" {
  description = "The IP address of the preferred file server"
  value       = aws_fsx_windows_file_system.this.preferred_file_server_ip
}

output "remote_administration_endpoint" {
  description = "The endpoint for remote PowerShell administration"
  value       = aws_fsx_windows_file_system.this.remote_administration_endpoint
}

output "network_interface_ids" {
  description = "The network interface IDs created for the file system"
  value       = aws_fsx_windows_file_system.this.network_interface_ids
}

output "vpc_id" {
  description = "The VPC ID in which the file system was created"
  value       = aws_fsx_windows_file_system.this.vpc_id
}

output "owner_id" {
  description = "The AWS account ID of the file system owner"
  value       = aws_fsx_windows_file_system.this.owner_id
}
