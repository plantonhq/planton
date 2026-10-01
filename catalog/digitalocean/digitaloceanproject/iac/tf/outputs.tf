# Stack outputs — exactly the DigitalOceanProjectStackOutputs contract,
# identical across both provisioners.

output "project_id" {
  description = "UUID of the project (the API identity, and the import id)"
  value       = digitalocean_project.project.id
}

output "owner_uuid" {
  description = "UUID of the account or team that owns the project"
  value       = digitalocean_project.project.owner_uuid
}

output "owner_id" {
  description = "Numeric id of the account or team that owns the project"
  value       = tostring(digitalocean_project.project.owner_id)
}

output "resource_urns" {
  description = "URNs of the resources DigitalOcean reports as project members after apply, sorted (the provider reads membership back whether or not the spec manages it)"
  value       = sort(tolist(digitalocean_project.project.resources))
}
