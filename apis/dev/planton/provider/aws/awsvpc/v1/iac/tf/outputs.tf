output "vpc_id" {
  description = "The ID of the VPC"
  value       = aws_vpc.this.id
}

output "vpc_arn" {
  description = "The ARN of the VPC"
  value       = aws_vpc.this.arn
}

output "cidr_block" {
  description = "The primary IPv4 CIDR block of the VPC"
  value       = aws_vpc.this.cidr_block
}

output "ipv6_cidr_block" {
  description = "The IPv6 CIDR block associated with the VPC (empty when IPv4-only)"
  value       = aws_vpc.this.ipv6_cidr_block
}

output "owner_id" {
  description = "The AWS account ID that owns the VPC"
  value       = aws_vpc.this.owner_id
}

output "main_route_table_id" {
  description = "The ID of the VPC's main route table"
  value       = aws_vpc.this.main_route_table_id
}

output "default_security_group_id" {
  description = "The ID of the default security group created with the VPC"
  value       = aws_vpc.this.default_security_group_id
}

output "default_network_acl_id" {
  description = "The ID of the default network ACL created with the VPC"
  value       = aws_vpc.this.default_network_acl_id
}

output "default_route_table_id" {
  description = "The ID of the default route table created with the VPC"
  value       = aws_vpc.this.default_route_table_id
}

output "region" {
  description = "The region the VPC was created in"
  value       = var.spec.region
}
