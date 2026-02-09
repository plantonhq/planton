resource "openstack_dns_recordset_v2" "main" {
  zone_id     = local.zone_id
  name        = local.record_name
  type        = local.record_type
  records     = var.spec.values
  ttl         = var.spec.ttl
  description = var.spec.description != "" ? var.spec.description : null
  region      = var.spec.region != "" ? var.spec.region : null
}
