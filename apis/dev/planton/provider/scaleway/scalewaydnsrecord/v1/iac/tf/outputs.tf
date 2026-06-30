# Stack outputs matching ScalewayDnsRecordStackOutputs proto.

output "record_id" {
  description = "The unique identifier of the created DNS record"
  value       = scaleway_domain_record.record.id
}

output "fqdn" {
  description = "The fully qualified domain name of the DNS record"
  value       = scaleway_domain_record.record.fqdn
}

# Complete outputs object matching stack_outputs.proto structure.
# Used by the Planton platform to populate status.outputs.
output "outputs" {
  description = "Complete DNS record outputs for integration with other resources"
  value = {
    record_id = scaleway_domain_record.record.id
    fqdn      = scaleway_domain_record.record.fqdn
  }
}
