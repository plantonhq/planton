resource "alicloud_cdn_domain_new" "main" {
  domain_name       = var.spec.domain_name
  cdn_type          = var.spec.cdn_type
  scope             = var.spec.scope != "" ? var.spec.scope : null
  check_url         = var.spec.check_url != "" ? var.spec.check_url : null
  resource_group_id = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
  tags              = local.final_tags

  dynamic "sources" {
    for_each = var.spec.sources
    content {
      type     = sources.value.type
      content  = sources.value.content
      port     = sources.value.port
      priority = sources.value.priority
      weight   = sources.value.weight
    }
  }

  dynamic "certificate_config" {
    for_each = var.spec.certificate_config != null ? [var.spec.certificate_config] : []
    content {
      cert_name                   = certificate_config.value.cert_name != "" ? certificate_config.value.cert_name : null
      cert_type                   = certificate_config.value.cert_type != "" ? certificate_config.value.cert_type : null
      cert_id                     = certificate_config.value.cert_id != "" ? certificate_config.value.cert_id : null
      cert_region                 = certificate_config.value.cert_region != "" ? certificate_config.value.cert_region : null
      server_certificate          = certificate_config.value.server_certificate != "" ? certificate_config.value.server_certificate : null
      private_key                 = certificate_config.value.private_key != "" ? certificate_config.value.private_key : null
      server_certificate_status   = certificate_config.value.server_certificate_status
    }
  }
}
