locals {
  standard_labels = merge(
    var.metadata.labels != null ? var.metadata.labels : {},
    {
      "planton-ai_resource" = "true"
      "planton-ai_name"     = var.metadata.name
      "planton_c_kind"      = "HetznerCloudDnsZone"
      "planton-ai_org"      = var.metadata.org != null ? var.metadata.org : ""
      "planton-ai_env"      = var.metadata.env != null ? var.metadata.env : ""
      "planton-ai_id"       = var.metadata.id != null ? var.metadata.id : ""
    },
  )

  record_sets = {
    for rs in (var.spec.record_sets != null ? var.spec.record_sets : []) :
    "${rs.name}-${lower(rs.type)}" => rs
  }
}
