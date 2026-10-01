locals {
  # Honor the spec contract: an empty project_id falls back to the
  # provider's default project.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # image_name falls back to metadata.name -- explicit conditional, so both
  # engines derive the identical cloud-side name.
  image_name = var.spec.image_name != "" ? var.spec.image_name : var.metadata.name

  # Empty optional strings become null so the provider applies its own
  # defaults instead of receiving an empty string it would reject or diff
  # on.
  description             = var.spec.description != "" ? var.spec.description : null
  family                  = var.spec.family != "" ? var.spec.family : null
  source_disk             = var.spec.source_disk != "" ? var.spec.source_disk : null
  source_image            = var.spec.source_image != "" ? var.spec.source_image : null
  source_snapshot         = var.spec.source_snapshot != "" ? var.spec.source_snapshot : null
  kms_key                 = var.spec.kms_key != "" ? var.spec.kms_key : null
  kms_key_service_account = var.spec.kms_key_service_account != "" ? var.spec.kms_key_service_account : null
  deletion_policy         = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # Optional+Computed lists: sent only when declared, so the values Google
  # inherits from the source never show as a diff.
  guest_os_features = length(var.spec.guest_os_features) > 0 ? var.spec.guest_os_features : []
  licenses          = length(var.spec.licenses) > 0 ? var.spec.licenses : null
  storage_locations = length(var.spec.storage_locations) > 0 ? var.spec.storage_locations : null

  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = local.image_name
    "planton-ai_kind"     = "gcpcomputeimage"
  }

  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "planton-ai_organization" = var.metadata.org } : {}

  env_label = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "planton-ai_environment" = var.metadata.env } : {}

  id_label = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "planton-ai_id" = var.metadata.id } : {}

  # User labels merge in first so the platform attribution labels can never
  # be clobbered by a spec label with the same key.
  final_labels = merge(var.spec.labels, local.base_labels, local.org_label, local.env_label, local.id_label)

  # The three source-decryption keys, each sent only with its source.
  source_encryptions = {
    source_disk_encryption_key     = var.spec.source_disk_encryption
    source_image_encryption_key    = var.spec.source_image_encryption
    source_snapshot_encryption_key = var.spec.source_snapshot_encryption
  }
}
