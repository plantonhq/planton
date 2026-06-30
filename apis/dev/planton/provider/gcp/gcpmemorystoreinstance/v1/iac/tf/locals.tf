locals {
  project_id    = var.spec.project_id.value
  instance_name = var.spec.instance_name
  location      = var.spec.location
  shard_count   = var.spec.shard_count

  mode           = var.spec.mode != "" ? var.spec.mode : null
  node_type      = var.spec.node_type != "" ? var.spec.node_type : null
  engine_version = var.spec.engine_version != "" ? var.spec.engine_version : null
  replica_count  = var.spec.replica_count > 0 ? var.spec.replica_count : null

  authorization_mode      = var.spec.authorization_mode != "" ? var.spec.authorization_mode : null
  transit_encryption_mode = var.spec.transit_encryption_mode != "" ? var.spec.transit_encryption_mode : null
  kms_key                 = var.spec.kms_key != null ? var.spec.kms_key.value : null

  deletion_protection_enabled = var.spec.deletion_protection_enabled

  labels = merge(
    {
      "planton-resource"      = "true"
      "planton-resource-name" = local.instance_name
      "planton-resource-kind" = "gcpmemorystoreinstance"
    },
    var.metadata.org != "" ? { "planton-organization" = var.metadata.org } : {},
    var.metadata.env != "" ? { "planton-environment" = var.metadata.env } : {},
    var.metadata.id != "" ? { "planton-resource-id" = var.metadata.id } : {},
  )
}
