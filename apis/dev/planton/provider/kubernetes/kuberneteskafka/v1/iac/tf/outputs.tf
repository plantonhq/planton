output "namespace" {
  description = "Namespace in which the Kafka resources are deployed."
  value       = local.namespace
}

output "username" {
  description = "The Kafka admin username."
  value       = local.admin_username
}

output "password_secret_name" {
  description = "The name of the Secret containing the Kafka admin SCRAM-SHA-512 credentials."
  value       = local.admin_password_secret_name
}

output "password_secret_key" {
  description = "The key within the Secret that contains the admin user's password."
  value       = "password"
}

output "bootstrap_server_external_hostname" {
  description = "External hostname for Kafka's bootstrap server (null if ingress is disabled)."
  value       = local.ingress_external_bootstrap_hostname
}

output "bootstrap_server_internal_hostname" {
  description = "Internal hostname for Kafka's bootstrap server (null if ingress is disabled)."
  value       = local.ingress_internal_bootstrap_hostname
}

output "schema_registry_external_url" {
  description = "External URL for the Schema Registry (including 'https://'), or null if not enabled/ingress disabled."
  value = (
    local.schema_registry_external_hostname != null
    ? "https://${local.schema_registry_external_hostname}"
    : null
  )
}

output "schema_registry_internal_url" {
  description = "Internal URL for the Schema Registry (including 'https://'), or null if not enabled/ingress disabled."
  value = (
    local.schema_registry_internal_hostname != null
    ? "https://${local.schema_registry_internal_hostname}"
    : null
  )
}

output "kafka_ui_external_url" {
  description = "External URL for the Kafka UI (Kowl), or null if it's not deployed or ingress is disabled."
  value = (
    local.kowl_external_hostname != null
    ? "https://${local.kowl_external_hostname}"
    : null
  )
}
