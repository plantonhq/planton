output "secret_id_map" {
  description = "Mapping from secret name to GCP secret ID"
  value       = {
    for secret_key, secret_res in google_secret_manager_secret.secrets :
    secret_key => secret_res.secret_id
  }
}
