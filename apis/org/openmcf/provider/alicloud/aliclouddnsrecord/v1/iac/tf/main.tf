resource "alicloud_alidns_record" "main" {
  domain_name = var.spec.domain_name
  rr          = var.spec.rr
  type        = var.spec.type
  value       = var.spec.value
  ttl         = var.spec.ttl
  priority    = var.spec.type == "MX" ? var.spec.priority : null
  line        = var.spec.line != "" ? var.spec.line : null
  status      = var.spec.status != "" ? var.spec.status : null
  remark      = var.spec.remark != "" ? var.spec.remark : null
}
