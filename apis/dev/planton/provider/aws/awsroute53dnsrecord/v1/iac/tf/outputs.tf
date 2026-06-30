output "fqdn" {
  description = "Fully qualified domain name of the created DNS record"
  value       = aws_route53_record.record.fqdn
}

output "record_type" {
  description = "DNS record type (A, AAAA, CNAME, etc.)"
  value       = aws_route53_record.record.type
}

output "zone_id" {
  description = "Route53 hosted zone ID where the record was created"
  value       = local.zone_id
}

output "is_alias" {
  description = "Whether this is an alias record"
  value       = local.is_alias
}

output "set_identifier" {
  description = "Set identifier for routing policies"
  value       = var.spec.set_identifier
}
