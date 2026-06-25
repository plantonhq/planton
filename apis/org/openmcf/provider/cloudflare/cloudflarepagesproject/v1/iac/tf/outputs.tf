output "project_name" {
  description = "The project name (downstream resources reference this value)."
  value       = cloudflare_pages_project.main.name
}

output "subdomain" {
  description = "The project's *.pages.dev subdomain."
  value       = cloudflare_pages_project.main.subdomain
}

output "domains" {
  description = "The custom domains attached to the project."
  value       = cloudflare_pages_project.main.domains
}

output "created_on" {
  description = "RFC3339 timestamp of when the project was created."
  value       = cloudflare_pages_project.main.created_on
}
