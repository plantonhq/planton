locals {
  node_pool_name = var.spec.name

  base_tags = {
    resource      = "true"
    resource_name = var.metadata.name
    resource_kind = "alicloudkubernetesnodepool"
  }

  id_tags = var.metadata.id != null ? { resource_id = var.metadata.id } : {}

  org_tags = var.metadata.org != null ? { organization = var.metadata.org } : {}

  env_tags = var.metadata.env != null ? { environment = var.metadata.env } : {}

  final_tags = merge(
    local.base_tags,
    local.id_tags,
    local.org_tags,
    local.env_tags,
    var.spec.tags,
  )

  system_disk = var.spec.system_disk != null ? var.spec.system_disk : {
    category          = "cloud_essd"
    size              = 120
    performance_level = ""
    encrypted         = false
    kms_key_id        = ""
  }
}
