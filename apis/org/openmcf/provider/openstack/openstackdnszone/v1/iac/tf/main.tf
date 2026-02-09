# main.tf

# Create the OpenStack Designate DNS zone.
resource "openstack_dns_zone_v2" "main" {
  name        = local.zone_name
  email       = var.spec.email != "" ? var.spec.email : null
  description = var.spec.description != "" ? var.spec.description : null
  ttl         = var.spec.ttl
  type        = var.spec.type != "" ? var.spec.type : null
  masters     = length(var.spec.masters) > 0 ? toset(var.spec.masters) : null
  region      = var.spec.region != "" ? var.spec.region : null
}

# Create inline DNS record sets.
# Each record is keyed by "record_type-record_name" for stable state management.
# Adding, removing, or reordering records only affects the specific record changed.
resource "openstack_dns_recordset_v2" "records" {
  for_each = local.records_map

  zone_id = openstack_dns_zone_v2.main.id
  name    = each.value.record_name
  type    = lookup(local.record_type_names, each.value.record_type, "A")
  records = each.value.values
  ttl     = each.value.ttl
  region  = var.spec.region != "" ? var.spec.region : null
}
