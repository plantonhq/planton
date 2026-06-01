locals {
  gateway_name       = var.metadata.name
  namespace          = var.spec.namespace
  gateway_class_name = var.spec.gateway_class_name

  labels = {
    "app.kubernetes.io/name"       = "gateway"
    "app.kubernetes.io/instance"   = var.metadata.name
    "app.kubernetes.io/managed-by" = "openmcf"
    "app.kubernetes.io/component"  = "gateway"
  }

  # Each listener is mapped to its camelCase CRD form, omitting unset optional
  # fields so upstream/controller defaults flow through. The nested tls and
  # allowedRoutes blocks are built with the same null-pruning approach.
  listeners = [
    for l in var.spec.listeners : merge(
      {
        name     = l.name
        port     = l.port
        protocol = l.protocol
      },
      l.hostname != null ? { hostname = l.hostname } : {},
      l.tls != null ? {
        tls = merge(
          l.tls.mode != null ? { mode = l.tls.mode } : {},
          l.tls.certificate_refs != null ? {
            certificateRefs = [
              for r in l.tls.certificate_refs : merge(
                { name = r.name },
                r.group != null ? { group = r.group } : {},
                r.kind != null ? { kind = r.kind } : {},
                r.namespace != null ? { namespace = r.namespace } : {},
              )
            ]
          } : {},
          l.tls.options != null ? { options = l.tls.options } : {},
        )
      } : {},
      l.allowed_routes != null ? {
        allowedRoutes = merge(
          l.allowed_routes.namespaces != null ? {
            namespaces = merge(
              l.allowed_routes.namespaces.from != null ? { from = l.allowed_routes.namespaces.from } : {},
              l.allowed_routes.namespaces.selector != null ? {
                selector = merge(
                  l.allowed_routes.namespaces.selector.match_labels != null ? { matchLabels = l.allowed_routes.namespaces.selector.match_labels } : {},
                  l.allowed_routes.namespaces.selector.match_expressions != null ? {
                    matchExpressions = [
                      for e in l.allowed_routes.namespaces.selector.match_expressions : merge(
                        {
                          key      = e.key
                          operator = e.operator
                        },
                        e.values != null ? { values = e.values } : {},
                      )
                    ]
                  } : {},
                )
              } : {},
            )
          } : {},
          l.allowed_routes.kinds != null ? {
            kinds = [
              for k in l.allowed_routes.kinds : merge(
                { kind = k.kind },
                k.group != null ? { group = k.group } : {},
              )
            ]
          } : {},
        )
      } : {},
    )
  ]

  # Optional requested addresses.
  addresses = var.spec.addresses != null ? [
    for a in var.spec.addresses : merge(
      a.type != null ? { type = a.type } : {},
      a.value != null ? { value = a.value } : {},
    )
  ] : null

  # Optional infrastructure attributes.
  infrastructure = var.spec.infrastructure != null ? merge(
    var.spec.infrastructure.labels != null ? { labels = var.spec.infrastructure.labels } : {},
    var.spec.infrastructure.annotations != null ? { annotations = var.spec.infrastructure.annotations } : {},
    var.spec.infrastructure.parameters_ref != null ? {
      parametersRef = {
        group = var.spec.infrastructure.parameters_ref.group
        kind  = var.spec.infrastructure.parameters_ref.kind
        name  = var.spec.infrastructure.parameters_ref.name
      }
    } : {},
  ) : null

  # Optional ListenerSet attachment policy.
  allowed_listeners = var.spec.allowed_listeners != null ? merge(
    var.spec.allowed_listeners.namespaces != null ? {
      namespaces = merge(
        var.spec.allowed_listeners.namespaces.from != null ? { from = var.spec.allowed_listeners.namespaces.from } : {},
        var.spec.allowed_listeners.namespaces.selector != null ? {
          selector = merge(
            var.spec.allowed_listeners.namespaces.selector.match_labels != null ? { matchLabels = var.spec.allowed_listeners.namespaces.selector.match_labels } : {},
            var.spec.allowed_listeners.namespaces.selector.match_expressions != null ? {
              matchExpressions = [
                for e in var.spec.allowed_listeners.namespaces.selector.match_expressions : merge(
                  {
                    key      = e.key
                    operator = e.operator
                  },
                  e.values != null ? { values = e.values } : {},
                )
              ]
            } : {},
          )
        } : {},
      )
    } : {},
  ) : null

  # Gateway-wide backend TLS (Gateway-as-client).
  backend_tls = (var.spec.tls != null && var.spec.tls.backend != null && var.spec.tls.backend.client_certificate_ref != null) ? {
    clientCertificateRef = merge(
      { name = var.spec.tls.backend.client_certificate_ref.name },
      var.spec.tls.backend.client_certificate_ref.group != null ? { group = var.spec.tls.backend.client_certificate_ref.group } : {},
      var.spec.tls.backend.client_certificate_ref.kind != null ? { kind = var.spec.tls.backend.client_certificate_ref.kind } : {},
      var.spec.tls.backend.client_certificate_ref.namespace != null ? { namespace = var.spec.tls.backend.client_certificate_ref.namespace } : {},
    )
  } : null

  # Gateway-wide frontend TLS (inbound client-certificate validation).
  frontend_tls = (var.spec.tls != null && var.spec.tls.frontend != null) ? merge(
    {
      default = merge(
        var.spec.tls.frontend.default.validation != null ? {
          validation = merge(
            { caCertificateRefs = [
              for r in var.spec.tls.frontend.default.validation.ca_certificate_refs : merge(
                {
                  group = r.group
                  kind  = r.kind
                  name  = r.name
                },
                r.namespace != null ? { namespace = r.namespace } : {},
              )
            ] },
            var.spec.tls.frontend.default.validation.mode != null ? { mode = var.spec.tls.frontend.default.validation.mode } : {},
          )
        } : {},
      )
    },
    var.spec.tls.frontend.per_port != null ? {
      perPort = [
        for p in var.spec.tls.frontend.per_port : {
          port = p.port
          tls = merge(
            p.tls.validation != null ? {
              validation = merge(
                { caCertificateRefs = [
                  for r in p.tls.validation.ca_certificate_refs : merge(
                    {
                      group = r.group
                      kind  = r.kind
                      name  = r.name
                    },
                    r.namespace != null ? { namespace = r.namespace } : {},
                  )
                ] },
                p.tls.validation.mode != null ? { mode = p.tls.validation.mode } : {},
              )
            } : {},
          )
        }
      ]
    } : {},
  ) : null

  gateway_tls = var.spec.tls != null ? merge(
    local.backend_tls != null ? { backend = local.backend_tls } : {},
    local.frontend_tls != null ? { frontend = local.frontend_tls } : {},
  ) : null

  # Assemble the full Gateway spec, omitting unset optional top-level blocks.
  gateway_spec = merge(
    {
      gatewayClassName = local.gateway_class_name
      listeners        = local.listeners
    },
    local.addresses != null ? { addresses = local.addresses } : {},
    local.infrastructure != null ? { infrastructure = local.infrastructure } : {},
    local.allowed_listeners != null ? { allowedListeners = local.allowed_listeners } : {},
    local.gateway_tls != null ? { tls = local.gateway_tls } : {},
  )
}
