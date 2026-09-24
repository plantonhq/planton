locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The bare pool ID Google's resource is keyed by. The spec's pool arrives
  # either as the pool's full resource path (a GcpPrivateCaPool reference
  # resolves to its name output) or as the bare ID; the last path segment is
  # the ID in both cases -- identical to the Pulumi module's PoolId.
  pool = element(split("/", var.spec.pool), length(split("/", var.spec.pool)) - 1)

  # The authority ID defaults to metadata.name -- the same fallback the
  # Pulumi module applies.
  certificate_authority_id = var.spec.certificate_authority_id != "" ? var.spec.certificate_authority_id : var.metadata.name

  # Deletion guard, honest by default: unset means true, identical to the
  # Pulumi module.
  deletion_protection = var.spec.deletion_protection != null ? var.spec.deletion_protection : true

  # Empty optional strings become null so the provider omits them.
  type               = var.spec.type != "" ? var.spec.type : null
  lifetime           = var.spec.lifetime != "" ? var.spec.lifetime : null
  pem_ca_certificate = var.spec.pem_ca_certificate != "" ? var.spec.pem_ca_certificate : null
  gcs_bucket         = var.spec.gcs_bucket != "" ? var.spec.gcs_bucket : null
  desired_state      = var.spec.desired_state != "" ? var.spec.desired_state : null
  deletion_policy    = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # The same planton-ai_* label set the Pulumi module applies, so the
  # authority is attributable to its Planton object regardless of the engine
  # that created it. User labels merge in first so the platform attribution
  # labels can never be clobbered by a spec label with the same key.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = var.metadata.name
    "planton-ai_kind"     = "gcpprivatecacertificateauthority"
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
