locals {
  reference_grant_name = var.metadata.name
  namespace            = var.spec.namespace

  labels = {
    "app.kubernetes.io/name"       = "referencegrant"
    "app.kubernetes.io/instance"   = var.metadata.name
    "app.kubernetes.io/managed-by" = "openmcf"
    "app.kubernetes.io/component"  = "referencegrant"
  }

  # Trusted sources (group / kind / namespace), proto snake_case mapped to Gateway
  # API camelCase, with the optional group pruned when empty so the upstream core
  # group default flows through. from entries are trust assertions about kinds,
  # not foreign keys; from[].namespace is the one cross-resource reference and is
  # wired (when OpenMCF-managed) via metadata.relationships (DD-009).
  from = [
    for f in var.spec.from : merge(
      {
        kind      = f.kind
        namespace = f.namespace
      },
      f.group != null && f.group != "" ? { group = f.group } : {},
    )
  ]

  # Referenceable targets (group / kind / optional name); group pruned when empty,
  # name pruned when unset (absence = all resources of the group/kind).
  to = [
    for t in var.spec.to : merge(
      { kind = t.kind },
      t.group != null && t.group != "" ? { group = t.group } : {},
      t.name != null ? { name = t.name } : {},
    )
  ]

  # Final ReferenceGrant spec manifest (camelCase). from and to are always present
  # (both required, min 1).
  reference_grant_spec = {
    from = local.from
    to   = local.to
  }
}
