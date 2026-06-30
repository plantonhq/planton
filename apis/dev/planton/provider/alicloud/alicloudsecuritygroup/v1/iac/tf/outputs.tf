output "security_group_id" {
  description = "The security group ID"
  value       = alicloud_security_group.main.id
}

output "security_group_name" {
  description = "The security group name"
  value       = alicloud_security_group.main.security_group_name
}
