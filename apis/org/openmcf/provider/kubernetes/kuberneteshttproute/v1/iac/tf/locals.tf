locals {
  route_name = var.metadata.name
  namespace  = var.spec.namespace

  labels = {
    "app.kubernetes.io/name"       = "httproute"
    "app.kubernetes.io/instance"   = var.metadata.name
    "app.kubernetes.io/managed-by" = "openmcf"
    "app.kubernetes.io/component"  = "httproute"
  }

  # Parent references (Gateways the route attaches to), proto snake_case mapped
  # to Gateway API camelCase, with unset optional fields pruned.
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

  # Routing rules. Each rule's matches, filters, backend refs, and timeouts are
  # built with null-pruning so only the fields the user set reach the manifest.
  rules = [
    for r in var.spec.rules : merge(
      r.name != null ? { name = r.name } : {},
      r.matches != null ? { matches = [
        for m in r.matches : merge(
          m.path != null ? { path = merge(
            m.path.type != null ? { type = m.path.type } : {},
            m.path.value != null ? { value = m.path.value } : {},
          ) } : {},
          m.headers != null ? { headers = [
            for h in m.headers : merge(
              { name = h.name, value = h.value },
              h.type != null ? { type = h.type } : {},
            )
          ] } : {},
          m.query_params != null ? { queryParams = [
            for q in m.query_params : merge(
              { name = q.name, value = q.value },
              q.type != null ? { type = q.type } : {},
            )
          ] } : {},
          m.method != null ? { method = m.method } : {},
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
          f.request_redirect != null ? { requestRedirect = merge(
            f.request_redirect.scheme != null ? { scheme = f.request_redirect.scheme } : {},
            f.request_redirect.hostname != null ? { hostname = f.request_redirect.hostname } : {},
            f.request_redirect.path != null ? { path = merge(
              { type = f.request_redirect.path.type },
              f.request_redirect.path.replace_full_path != null ? { replaceFullPath = f.request_redirect.path.replace_full_path } : {},
              f.request_redirect.path.replace_prefix_match != null ? { replacePrefixMatch = f.request_redirect.path.replace_prefix_match } : {},
            ) } : {},
            f.request_redirect.port != null ? { port = f.request_redirect.port } : {},
            f.request_redirect.status_code != null ? { statusCode = f.request_redirect.status_code } : {},
          ) } : {},
          f.url_rewrite != null ? { urlRewrite = merge(
            f.url_rewrite.hostname != null ? { hostname = f.url_rewrite.hostname } : {},
            f.url_rewrite.path != null ? { path = merge(
              { type = f.url_rewrite.path.type },
              f.url_rewrite.path.replace_full_path != null ? { replaceFullPath = f.url_rewrite.path.replace_full_path } : {},
              f.url_rewrite.path.replace_prefix_match != null ? { replacePrefixMatch = f.url_rewrite.path.replace_prefix_match } : {},
            ) } : {},
          ) } : {},
          f.cors != null ? { cors = merge(
            f.cors.allow_origins != null ? { allowOrigins = f.cors.allow_origins } : {},
            f.cors.allow_credentials != null ? { allowCredentials = f.cors.allow_credentials } : {},
            f.cors.allow_methods != null ? { allowMethods = f.cors.allow_methods } : {},
            f.cors.allow_headers != null ? { allowHeaders = f.cors.allow_headers } : {},
            f.cors.expose_headers != null ? { exposeHeaders = f.cors.expose_headers } : {},
            f.cors.max_age != null ? { maxAge = f.cors.max_age } : {},
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
              f.request_redirect != null ? { requestRedirect = merge(
                f.request_redirect.scheme != null ? { scheme = f.request_redirect.scheme } : {},
                f.request_redirect.hostname != null ? { hostname = f.request_redirect.hostname } : {},
                f.request_redirect.path != null ? { path = merge(
                  { type = f.request_redirect.path.type },
                  f.request_redirect.path.replace_full_path != null ? { replaceFullPath = f.request_redirect.path.replace_full_path } : {},
                  f.request_redirect.path.replace_prefix_match != null ? { replacePrefixMatch = f.request_redirect.path.replace_prefix_match } : {},
                ) } : {},
                f.request_redirect.port != null ? { port = f.request_redirect.port } : {},
                f.request_redirect.status_code != null ? { statusCode = f.request_redirect.status_code } : {},
              ) } : {},
              f.url_rewrite != null ? { urlRewrite = merge(
                f.url_rewrite.hostname != null ? { hostname = f.url_rewrite.hostname } : {},
                f.url_rewrite.path != null ? { path = merge(
                  { type = f.url_rewrite.path.type },
                  f.url_rewrite.path.replace_full_path != null ? { replaceFullPath = f.url_rewrite.path.replace_full_path } : {},
                  f.url_rewrite.path.replace_prefix_match != null ? { replacePrefixMatch = f.url_rewrite.path.replace_prefix_match } : {},
                ) } : {},
              ) } : {},
              f.cors != null ? { cors = merge(
                f.cors.allow_origins != null ? { allowOrigins = f.cors.allow_origins } : {},
                f.cors.allow_credentials != null ? { allowCredentials = f.cors.allow_credentials } : {},
                f.cors.allow_methods != null ? { allowMethods = f.cors.allow_methods } : {},
                f.cors.allow_headers != null ? { allowHeaders = f.cors.allow_headers } : {},
                f.cors.expose_headers != null ? { exposeHeaders = f.cors.expose_headers } : {},
                f.cors.max_age != null ? { maxAge = f.cors.max_age } : {},
              ) } : {},
              f.extension_ref != null ? { extensionRef = { group = f.extension_ref.group, kind = f.extension_ref.kind, name = f.extension_ref.name } } : {},
            )
          ] } : {},
        )
      ] } : {},
      r.timeouts != null ? { timeouts = merge(
        r.timeouts.request != null ? { request = r.timeouts.request } : {},
        r.timeouts.backend_request != null ? { backendRequest = r.timeouts.backend_request } : {},
      ) } : {},
    )
  ]

  # Final HTTPRoute spec manifest (camelCase), with optional top-level blocks
  # pruned when unset.
  http_route_spec = merge(
    { rules = local.rules },
    local.parent_refs != null ? { parentRefs = local.parent_refs } : {},
    var.spec.hostnames != null ? { hostnames = var.spec.hostnames } : {},
  )
}
