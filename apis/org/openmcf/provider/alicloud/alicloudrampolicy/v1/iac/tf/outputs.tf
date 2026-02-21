output "policy_name" {
  description = "The RAM policy name"
  value       = alicloud_ram_policy.main.policy_name
}

output "policy_type" {
  description = "The RAM policy type (always Custom for user-created policies)"
  value       = alicloud_ram_policy.main.type
}
