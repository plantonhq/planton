# Create the public DNS zone -- an internet-facing, authoritative DNS zone
# hosted on Azure's global anycast name-server fleet.
#
# Lifecycle notes worth knowing before operating this resource:
# - Tags update IN PLACE; the zone's name is its ARM identity, so renaming
#   replaces the zone, every record in it, AND the assigned name-server
#   set -- breaking the registrar delegation until it is updated.
# - Public DNS zones are global resources -- no location/region argument.
# - The zone is deliberately just the zone: records are declared through
#   AzureDnsRecord resources referencing this zone's zone_name output, one
#   resource per record set.
# - Creating the zone does NOT make it authoritative: the domain resolves
#   through it only once the nameservers output is configured at the
#   registrar (or as parent-zone NS records for subdomain delegation).
resource "azurerm_dns_zone" "main" {
  name                = var.spec.zone_name
  resource_group_name = var.spec.resource_group

  # Absent means Azure creates its standard SOA record; the block is only
  # sent when the spec customizes it. Unset timers fall back to Azure's
  # defaults so a partially-specified block and Azure's own values deploy
  # identically on both engines. The SOA host name is never sent -- Azure
  # owns it and rejects changes.
  dynamic "soa_record" {
    for_each = var.spec.soa_record != null ? [var.spec.soa_record] : []
    content {
      email         = soa_record.value.email
      expire_time   = soa_record.value.expire_time
      minimum_ttl   = soa_record.value.minimum_ttl
      refresh_time  = soa_record.value.refresh_time
      retry_time    = soa_record.value.retry_time
      serial_number = soa_record.value.serial_number
      ttl           = soa_record.value.ttl
      tags          = soa_record.value.tags
    }
  }

  tags = local.final_tags
}
