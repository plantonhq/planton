locals {
  certificate_name = var.metadata.name
  is_uploaded      = var.spec.uploaded != null
  is_managed       = var.spec.managed != null

  standard_labels = merge(
    var.metadata.labels != null ? var.metadata.labels : {},
    {
      "planton-ai_resource" = "true"
      "planton-ai_name"     = var.metadata.name
      "planton-ai_kind"     = "HetznerCloudCertificate"
      "planton-ai_org"      = var.metadata.org != null ? var.metadata.org : ""
      "planton-ai_env"      = var.metadata.env != null ? var.metadata.env : ""
      "planton-ai_id"       = var.metadata.id != null ? var.metadata.id : ""
    },
  )
}
