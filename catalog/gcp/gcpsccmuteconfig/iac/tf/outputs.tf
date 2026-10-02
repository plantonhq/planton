output "name" {
  description = "Full resource name of the mute config"
  value = one(concat(
    google_scc_v2_project_mute_config.this[*].name,
    google_scc_v2_folder_mute_config.this[*].name,
    google_scc_v2_organization_mute_config.this[*].name,
  ))
}
