# outputs.tf

output "zone_id" {
  description = "The unique identifier (UUID) of the DNS zone"
  value       = openstack_dns_zone_v2.main.id
}

output "zone_name" {
  description = "The DNS zone name"
  value       = openstack_dns_zone_v2.main.name
}

output "region" {
  description = "The OpenStack region where the DNS zone was created"
  value       = openstack_dns_zone_v2.main.region
}
