output "name" {
  description = "Full resource name of the BigQuery export"
  value = one(concat(
    google_scc_v2_project_scc_big_query_export.this[*].name,
    google_scc_v2_folder_scc_big_query_export.this[*].name,
    google_scc_v2_organization_scc_big_query_export.this[*].name,
  ))
}

output "principal" {
  description = "The Security Command Center service account that writes the findings (grant it roles/bigquery.dataEditor on the dataset)"
  value = one(concat(
    google_scc_v2_project_scc_big_query_export.this[*].principal,
    google_scc_v2_folder_scc_big_query_export.this[*].principal,
    google_scc_v2_organization_scc_big_query_export.this[*].principal,
  ))
}
