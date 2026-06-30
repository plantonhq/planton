output "project_name" {
  description = "The SLS project name"
  value       = alicloud_log_project.main.project_name
}

output "project_id" {
  description = "The SLS project ID"
  value       = alicloud_log_project.main.id
}

output "log_store_names" {
  description = "Map of log store names created within the project"
  value       = { for name, store in alicloud_log_store.stores : name => store.logstore_name }
}
