locals {
  route_name = var.metadata.name
  namespace  = var.spec.namespace

  labels = {
    "app.kubernetes.io/name"       = "grpcroute"
    "app.kubernetes.io/instance"   = var.metadata.name
    "app.kubernetes.io/managed-by" = "openmcf"
    "app.kubernetes.io/component"  = "grpcroute"
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

  # Routing rules. Each rule's matches, filters, and backend refs are built with
  # null-pruning so only the fields the user set reach the manifest. GRPCRouteRule
  # has no timeouts.
  rules = [
    for r in var.spec.rules : merge(
      r.name != null ? { name = r.name } : {},
      r.matches != null ? { matches = [
        for m in r.matches : merge(
          m.method != null ? { method = merge(
            m.method.type != null ? { type = m.method.type } : {},
            m.method.service != null ? { service = m.method.service } : {},
            m.method.method != null ? { method = m.method.method } : {},
          ) } : {},
          m.headers != null ? { headers = [
            for h in m.headers : merge(
              { name = h.name, value = h.value },
              h.type != null ? { type = h.type } : {},
            )
          ] } : {},
        )
      ] } : {},
      r.filters != null ? { filters = [
        for f in r.filters : merge(
          { type = f.type },
          f.request_header_modifier != null ? { requestHeaderModifier = merge(
            f.request_header_modifier.set != null ? { set = [for h in f.request_header_modifier.set : { name = h.name, value = h.value }] } : {},
            f.request_header_modifier.add != null ? { add = [for h in f.request_header_modifier.add : { name = h.name, value = h.value }] } : {},
            f.request_header_modifier.remove != null ? { remove = f.request_header_modifier.remove } : {},
          ) } : {},
          f.response_header_modifier != null ? { responseHeaderModifier = merge(
            f.response_header_modifier.set != null ? { set = [for h in f.response_header_modifier.set : { name = h.name, value = h.value }] } : {},
            f.response_header_modifier.add != null ? { add = [for h in f.response_header_modifier.add : { name = h.name, value = h.value }] } : {},
            f.response_header_modifier.remove != null ? { remove = f.response_header_modifier.remove } : {},
          ) } : {},
          f.request_mirror != null ? { requestMirror = merge(
            { backendRef = merge(
              { name = f.request_mirror.backend_ref.name },
              f.request_mirror.backend_ref.group != null ? { group = f.request_mirror.backend_ref.group } : {},
              f.request_mirror.backend_ref.kind != null ? { kind = f.request_mirror.backend_ref.kind } : {},
              f.request_mirror.backend_ref.namespace != null ? { namespace = f.request_mirror.backend_ref.namespace } : {},
              f.request_mirror.backend_ref.port != null ? { port = f.request_mirror.backend_ref.port } : {},
            ) },
            f.request_mirror.percent != null ? { percent = f.request_mirror.percent } : {},
            f.request_mirror.fraction != null ? { fraction = merge(
              { numerator = f.request_mirror.fraction.numerator },
              f.request_mirror.fraction.denominator != null ? { denominator = f.request_mirror.fraction.denominator } : {},
            ) } : {},
          ) } : {},
          f.extension_ref != null ? { extensionRef = { group = f.extension_ref.group, kind = f.extension_ref.kind, name = f.extension_ref.name } } : {},
        )
      ] } : {},
      r.backend_refs != null ? { backendRefs = [
        for b in r.backend_refs : merge(
          { name = b.name },
          b.group != null ? { group = b.group } : {},
          b.kind != null ? { kind = b.kind } : {},
          b.namespace != null ? { namespace = b.namespace } : {},
          b.port != null ? { port = b.port } : {},
          b.weight != null ? { weight = b.weight } : {},
          b.filters != null ? { filters = [
            for f in b.filters : merge(
              { type = f.type },
              f.request_header_modifier != null ? { requestHeaderModifier = merge(
                f.request_header_modifier.set != null ? { set = [for h in f.request_header_modifier.set : { name = h.name, value = h.value }] } : {},
                f.request_header_modifier.add != null ? { add = [for h in f.request_header_modifier.add : { name = h.name, value = h.value }] } : {},
                f.request_header_modifier.remove != null ? { remove = f.request_header_modifier.remove } : {},
              ) } : {},
              f.response_header_modifier != null ? { responseHeaderModifier = merge(
                f.response_header_modifier.set != null ? { set = [for h in f.response_header_modifier.set : { name = h.name, value = h.value }] } : {},
                f.response_header_modifier.add != null ? { add = [for h in f.response_header_modifier.add : { name = h.name, value = h.value }] } : {},
                f.response_header_modifier.remove != null ? { remove = f.response_header_modifier.remove } : {},
              ) } : {},
              f.request_mirror != null ? { requestMirror = merge(
                { backendRef = merge(
                  { name = f.request_mirror.backend_ref.name },
                  f.request_mirror.backend_ref.group != null ? { group = f.request_mirror.backend_ref.group } : {},
                  f.request_mirror.backend_ref.kind != null ? { kind = f.request_mirror.backend_ref.kind } : {},
                  f.request_mirror.backend_ref.namespace != null ? { namespace = f.request_mirror.backend_ref.namespace } : {},
                  f.request_mirror.backend_ref.port != null ? { port = f.request_mirror.backend_ref.port } : {},
                ) },
                f.request_mirror.percent != null ? { percent = f.request_mirror.percent } : {},
                f.request_mirror.fraction != null ? { fraction = merge(
                  { numerator = f.request_mirror.fraction.numerator },
                  f.request_mirror.fraction.denominator != null ? { denominator = f.request_mirror.fraction.denominator } : {},
                ) } : {},
              ) } : {},
              f.extension_ref != null ? { extensionRef = { group = f.extension_ref.group, kind = f.extension_ref.kind, name = f.extension_ref.name } } : {},
            )
          ] } : {},
        )
      ] } : {},
    )
  ]

  # Final GRPCRoute spec manifest (camelCase), with optional top-level blocks
  # pruned when unset.
  grpc_route_spec = merge(
    { rules = local.rules },
    local.parent_refs != null ? { parentRefs = local.parent_refs } : {},
    var.spec.hostnames != null ? { hostnames = var.spec.hostnames } : {},
  )
}
