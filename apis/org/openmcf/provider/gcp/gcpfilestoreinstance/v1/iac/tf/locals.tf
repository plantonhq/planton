locals {
  project_id    = var.spec.project_id.value
  instance_name = var.spec.instance_name
  location      = var.spec.location
  tier          = var.spec.tier

  description  = var.spec.description != "" ? var.spec.description : null
  protocol     = var.spec.protocol != "" ? var.spec.protocol : null
  kms_key_name = var.spec.kms_key_name != null ? var.spec.kms_key_name.value : null

  network           = var.spec.network_config.network.value
  connect_mode      = var.spec.network_config.connect_mode != "" ? var.spec.network_config.connect_mode : null
  reserved_ip_range = var.spec.network_config.reserved_ip_range != "" ? var.spec.network_config.reserved_ip_range : null

  labels = merge(
    {
      "openmcf-resource"      = "true"
      "openmcf-resource-name" = local.instance_name
      "openmcf-resource-kind" = "gcpfilestoreinstance"
    },
    var.metadata.org != "" ? { "openmcf-organization" = var.metadata.org } : {},
    var.metadata.env != "" ? { "openmcf-environment" = var.metadata.env } : {},
    var.metadata.id != "" ? { "openmcf-resource-id" = var.metadata.id } : {},
  )
}
