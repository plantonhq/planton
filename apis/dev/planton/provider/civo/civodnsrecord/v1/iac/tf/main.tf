# main.tf

# Create the Civo DNS Record
resource "civo_dns_domain_record" "main" {
  domain_id = var.spec.zone_id
  name      = var.spec.name
  type      = local.record_type
  value     = var.spec.value
  ttl       = local.ttl

  # Priority is only used for MX and SRV records
  priority = local.requires_priority ? var.spec.priority : null
}
