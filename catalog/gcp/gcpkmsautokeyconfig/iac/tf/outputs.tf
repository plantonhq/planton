output "name" {
  description = "Full resource name of the Autokey configuration (folders/{id}/autokeyConfig or projects/{id}/autokeyConfig)"
  value       = one(concat(google_kms_autokey_config.this[*].id, google_kms_project_autokey_config.this[*].id))
}

output "parent" {
  description = "The folder or project Autokey is configured on"
  value       = local.is_folder ? "folders/${local.folder_id}" : "projects/${local.config_project}"
}
