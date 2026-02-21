output "role_id" {
  description = "The RAM role ID"
  value       = alicloud_ram_role.main.role_id
}

output "role_name" {
  description = "The RAM role name"
  value       = alicloud_ram_role.main.role_name
}

output "arn" {
  description = "The RAM role ARN"
  value       = alicloud_ram_role.main.arn
}
