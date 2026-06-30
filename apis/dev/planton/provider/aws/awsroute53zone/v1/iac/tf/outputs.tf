###############################################################################
# Outputs mapped to AwsRoute53ZoneStackOutputs
###############################################################################

output "zone_id" {
  description = "The hosted zone ID"
  value       = aws_route53_zone.r53_zone.zone_id
}

output "zone_name" {
  description = "The hosted zone name"
  value       = aws_route53_zone.r53_zone.name
}

output "nameservers" {
  description = "The nameservers for the hosted zone"
  value       = aws_route53_zone.r53_zone.name_servers
}

output "caller_reference" {
  description = "The caller reference used when creating the zone"
  value       = ""
}


