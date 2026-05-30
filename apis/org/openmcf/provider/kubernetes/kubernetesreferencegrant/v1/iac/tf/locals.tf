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
  # API camelCase. All three are required by the upstream CRD: group is a
  # non-pointer value type (no omitempty), so the API server rejects the resource
  # unless the key is present -- even for the core API group, where the value is
  # the empty string. We therefore always emit group (null normalized to "").
  # from entries are trust assertions about kinds, not foreign keys;
  # from[].namespace is the one cross-resource reference and is wired (when
  # OpenMCF-managed) via metadata.relationships (DD-009).
  from = [
    for f in var.spec.from : {
      group     = f.group != null ? f.group : ""
      kind      = f.kind
      namespace = f.namespace
    }
  ]

  # Referenceable targets (group / kind / optional name). group and kind are
  # required by the CRD, so group is always emitted (empty string = core API
  # group; the key must still be present). name is the only optional field and is
  # pruned when unset (absence = all resources of the group/kind).
  to = [
    for t in var.spec.to : merge(
      {
        group = t.group != null ? t.group : ""
        kind  = t.kind
      },
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
