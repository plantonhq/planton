locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # Bare IDs Google's resource is keyed by. The pool and the signing
  # authority arrive either as full resource paths (references resolve to
  # their name outputs) or as bare IDs; the last path segment is the ID in
  # both cases -- identical to the Pulumi module.
  pool                  = element(split("/", var.spec.pool), length(split("/", var.spec.pool)) - 1)
  certificate_authority = var.spec.certificate_authority != "" ? element(split("/", var.spec.certificate_authority), length(split("/", var.spec.certificate_authority)) - 1) : null

  # The certificate ID defaults to metadata.name -- the same fallback the
  # Pulumi module applies.
  certificate_id = var.spec.certificate_id != "" ? var.spec.certificate_id : var.metadata.name

  # Empty optional strings become null so the provider omits them.
  certificate_template = var.spec.certificate_template != "" ? var.spec.certificate_template : null
  lifetime             = var.spec.lifetime != "" ? var.spec.lifetime : null
  pem_csr              = var.spec.pem_csr != "" ? var.spec.pem_csr : null
  deletion_policy      = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # The same planton-ai_* label set the Pulumi module applies, so the
  # certificate is attributable to its Planton object regardless of the engine
  # that created it. User labels merge in first so the platform attribution
  # labels can never be clobbered by a spec label with the same key.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = var.metadata.name
    "planton-ai_kind"     = "gcpprivatecacertificate"
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

  final_labels = merge(var.spec.labels, local.base_labels, local.org_label, local.env_label, local.id_label)
}
