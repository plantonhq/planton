locals {
  project_id    = var.spec.project_id.value
  instance_name = var.spec.instance_name
  config        = var.spec.config
  display_name  = var.spec.display_name

  num_nodes        = var.spec.num_nodes > 0 ? var.spec.num_nodes : null
  processing_units = var.spec.processing_units > 0 ? var.spec.processing_units : null

  instance_type                = var.spec.instance_type != "" ? var.spec.instance_type : null
  edition                      = var.spec.edition != "" ? var.spec.edition : null
  default_backup_schedule_type = var.spec.default_backup_schedule_type != "" ? var.spec.default_backup_schedule_type : null

  labels = merge(
    {
      "openmcf-resource"      = "true"
      "openmcf-resource-name" = local.instance_name
      "openmcf-resource-kind" = "gcpspannerinstance"
    },
    var.metadata.org != "" ? { "openmcf-organization" = var.metadata.org } : {},
    var.metadata.env != "" ? { "openmcf-environment" = var.metadata.env } : {},
    var.metadata.id != "" ? { "openmcf-resource-id" = var.metadata.id } : {},
  )
}
