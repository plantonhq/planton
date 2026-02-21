locals {
  placement_group_name = var.metadata.name
  placement_group_type = coalesce(var.spec.type, "spread")

  standard_labels = merge(
    var.metadata.labels != null ? var.metadata.labels : {},
    {
      "planton-ai_resource" = "true"
      "planton-ai_name"     = var.metadata.name
      "planton-ai_kind"     = "HetznerCloudPlacementGroup"
      "planton-ai_org"      = var.metadata.org != null ? var.metadata.org : ""
      "planton-ai_env"      = var.metadata.env != null ? var.metadata.env : ""
      "planton-ai_id"       = var.metadata.id != null ? var.metadata.id : ""
    },
  )
}
