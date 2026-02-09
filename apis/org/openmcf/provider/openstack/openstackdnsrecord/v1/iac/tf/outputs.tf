output "recordset_id" {
  description = "The unique identifier (UUID) of the DNS recordset"
  value       = openstack_dns_recordset_v2.main.id
}

output "fqdn" {
  description = "The fully qualified domain name of the record"
  value       = openstack_dns_recordset_v2.main.name
}

output "record_type" {
  description = "The DNS record type"
  value       = openstack_dns_recordset_v2.main.type
}

output "values" {
  description = "The DNS record values"
  value       = openstack_dns_recordset_v2.main.records
}

output "zone_id" {
  description = "The zone ID containing this record"
  value       = openstack_dns_recordset_v2.main.zone_id
}

output "region" {
  description = "The OpenStack region"
  value       = openstack_dns_recordset_v2.main.region
}
