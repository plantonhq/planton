# Security Group ID
output "security_group_id" {
  description = "The unique identifier of the created Scaleway Instance Security Group"
  value       = scaleway_instance_security_group.sg.id
}

# Organization ID
output "organization_id" {
  description = "The Organization ID the security group is associated with"
  value       = scaleway_instance_security_group.sg.organization_id
}
