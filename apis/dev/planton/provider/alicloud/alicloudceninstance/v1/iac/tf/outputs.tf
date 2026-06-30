output "cen_id" {
  description = "The CEN instance ID"
  value       = alicloud_cen_instance.main.id
}

output "cen_instance_name" {
  description = "The CEN instance name"
  value       = alicloud_cen_instance.main.cen_instance_name
}
