locals {
  route_name = var.metadata.name
  namespace  = var.spec.namespace

  labels = {
    "app.kubernetes.io/name"       = "tcproute"
    "app.kubernetes.io/instance"   = var.metadata.name
    "app.kubernetes.io/managed-by" = "openmcf"
    "app.kubernetes.io/component"  = "tcproute"
  }

  # Parent references (Gateways the route attaches to), proto snake_case mapped
  # to Gateway API camelCase, with unset optional fields pruned. parent_refs is a
  # plain reference (DD-009); infra-chart authors wire the route -> Gateway edge
  # via metadata.relationships.
  parent_refs = var.spec.parent_refs != null ? [
    for p in var.spec.parent_refs : merge(
      { name = p.name },
      p.group != null ? { group = p.group } : {},
      p.kind != null ? { kind = p.kind } : {},
      p.namespace != null ? { namespace = p.namespace } : {},
      p.section_name != null ? { sectionName = p.section_name } : {},
      p.port != null ? { port = p.port } : {},
    )
  ] : null

  # Routing rules. A TCP route rule has only an optional name and the backend
  # refs (no matches, no filters); backend refs are null-pruned. backend_refs is
  # a plain reference (DD-009); infra-chart authors wire the route -> backend
  # edge via metadata.relationships.
  rules = [
    for r in var.spec.rules : merge(
      r.name != null ? { name = r.name } : {},
      {
        backendRefs = [
          for b in r.backend_refs : merge(
            { name = b.name },
            b.group != null ? { group = b.group } : {},
            b.kind != null ? { kind = b.kind } : {},
            b.namespace != null ? { namespace = b.namespace } : {},
            b.port != null ? { port = b.port } : {},
            b.weight != null ? { weight = b.weight } : {},
          )
        ]
      },
    )
  ]

  # Final TCPRoute spec manifest (camelCase). rules is always present (required);
  # parentRefs and useDefaultGateways are pruned when unset.
  tcp_route_spec = merge(
    { rules = local.rules },
    local.parent_refs != null ? { parentRefs = local.parent_refs } : {},
    var.spec.use_default_gateways != null ? { useDefaultGateways = var.spec.use_default_gateways } : {},
  )
}
