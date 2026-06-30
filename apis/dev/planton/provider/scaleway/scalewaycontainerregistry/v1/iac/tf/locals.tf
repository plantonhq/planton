locals {
  # ── Resource identity ──────────────────────────────────────────────
  namespace_name = var.metadata.name
  region         = var.spec.region

  # ── Configuration ──────────────────────────────────────────────────
  description = var.spec.description
  is_public   = var.spec.is_public

  # NOTE: Scaleway Container Registry namespaces do not support tags.
  # Unlike most other Scaleway resources, the registry API does not
  # accept tags/labels. Standard Planton metadata tags are not applied
  # to this resource. The namespace name (from metadata.name) is the
  # primary identifier for resource tracking and cost allocation.
}
