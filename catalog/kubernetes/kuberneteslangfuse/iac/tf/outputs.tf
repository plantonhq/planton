# Outputs - identical names and derivations in the Pulumi module
# (KubernetesLangfuseOutputs). The seeding outputs are empty when seeding is
# off, and the admin outputs when no admin is seeded.

output "namespace" {
  description = "Namespace Langfuse runs in"
  value       = local.namespace
}

output "service" {
  description = "The web Service (<metadata.name>-web) - the backend a route or ingress exposes"
  value       = local.web_service_name
}

output "endpoint" {
  description = "In-cluster base URL of the web server - the LANGFUSE_BASE_URL an in-cluster service's SDK uses"
  value       = local.endpoint
}

output "external_url" {
  description = "The public URL Langfuse was configured with"
  value       = local.external_url
}

output "otel_endpoint" {
  description = "In-cluster OTLP/HTTP traces endpoint (Basic auth with the project key pair)"
  value       = local.otel_endpoint
}

output "port_forward_command" {
  description = "Copy-paste command for reaching the web UI from a workstation"
  value       = local.port_forward_command
}

output "instance_secret_name" {
  description = "The module-owned Secret holding every secret value this resource manages"
  value       = local.instance_secret_name
}

output "organization_id" {
  description = "ID of the organization seeded at first start"
  value       = local.organization_id
}

output "project_id" {
  description = "ID of the project seeded at first start"
  value       = local.project_id
}

output "project_public_key" {
  description = "Public key of the seeded project key pair"
  value = !local.init_enabled ? "" : (
    local.declared_keys ? try(var.spec.init.project.public_key, "") : try(nonsensitive("pk-lf-${random_password.init_project_public_key[0].result}"), "")
  )
}

output "project_secret_key_secret" {
  description = "Where the seeded project secret key lives - a key of the instance Secret"
  value = {
    name = local.init_enabled ? local.instance_secret_name : ""
    key  = local.init_enabled ? "init-project-secret-key" : ""
  }
}

output "admin_email" {
  description = "Email of the seeded admin user"
  value       = local.admin_email
}

output "admin_password_secret" {
  description = "Where the seeded admin user's password lives - a key of the instance Secret"
  value = {
    name = local.admin_seeded ? local.instance_secret_name : ""
    key  = local.admin_seeded ? "init-user-password" : ""
  }
}
