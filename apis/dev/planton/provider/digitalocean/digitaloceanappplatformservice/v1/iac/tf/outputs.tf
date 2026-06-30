# DigitalOcean App Platform Service Outputs

output "app_id" {
  description = "The ID of the created App Platform application"
  value       = digitalocean_app.main.id
}

output "live_url" {
  description = "The live URL of the application (default ondigitalocean.app domain)"
  value       = digitalocean_app.main.live_url
}

output "default_ingress" {
  description = "The default ingress URL for the application"
  value       = digitalocean_app.main.default_ingress
}

output "app_urn" {
  description = "The uniform resource name (URN) of the app"
  value       = digitalocean_app.main.urn
}

output "created_at" {
  description = "The date and time when the app was created"
  value       = digitalocean_app.main.created_at
}

output "updated_at" {
  description = "The date and time when the app was last updated"
  value       = digitalocean_app.main.updated_at
}

output "active_deployment_id" {
  description = "The ID of the currently active deployment"
  value       = digitalocean_app.main.active_deployment_id
}

