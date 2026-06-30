output "app_id" {
  description = "The SAE application ID"
  value       = alicloud_sae_application.main.id
}

output "app_name" {
  description = "The SAE application name"
  value       = alicloud_sae_application.main.app_name
}
