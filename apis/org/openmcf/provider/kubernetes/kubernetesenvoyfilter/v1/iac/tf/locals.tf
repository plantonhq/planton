locals {
  envoy_filter_name = var.metadata.name
  namespace         = var.spec.namespace

  labels = {
    "app.kubernetes.io/name"       = "envoy-filter"
    "app.kubernetes.io/instance"   = var.metadata.name
    "app.kubernetes.io/managed-by" = "openmcf"
    "app.kubernetes.io/component"  = "envoy-filter"
  }

  # Workload selector, mapped to its camelCase CRD form (labels) and omitted entirely when no
  # labels are provided. The nested labels is read through a ?: conditional (which only
  # evaluates the taken branch) so a null selector never triggers an attribute access on null.
  workload_selector_labels = var.spec.workload_selector != null ? var.spec.workload_selector.labels : null
  workload_selector = local.workload_selector_labels != null ? {
    labels = local.workload_selector_labels
  } : null

  # Target refs, snake_case mapped to camelCase, with unset optional fields pruned per element.
  target_refs = var.spec.target_refs != null ? [
    for ref in var.spec.target_refs : merge(
      ref.group != null ? { group = ref.group } : {},
      ref.kind != null ? { kind = ref.kind } : {},
      ref.name != null ? { name = ref.name } : {},
      ref.namespace != null ? { namespace = ref.namespace } : {},
    )
  ] : null

  # Config patches, snake_case mapped to camelCase, with unset optional blocks pruned at every
  # nesting level. Each nested optional block is read through a ?: conditional (HCL && does not
  # short-circuit) so a null parent never triggers an attribute access on null.
  config_patches = var.spec.config_patches != null ? [
    for cp in var.spec.config_patches : merge(
      cp.apply_to != null ? { applyTo = cp.apply_to } : {},
      cp.match != null ? { match = merge(
        cp.match.context != null ? { context = cp.match.context } : {},
        cp.match.proxy != null ? { proxy = merge(
          cp.match.proxy.proxy_version != null ? { proxyVersion = cp.match.proxy.proxy_version } : {},
          cp.match.proxy.metadata != null ? { metadata = cp.match.proxy.metadata } : {},
        ) } : {},
        cp.match.listener != null ? { listener = merge(
          cp.match.listener.port_number != null ? { portNumber = cp.match.listener.port_number } : {},
          cp.match.listener.listener_filter != null ? { listenerFilter = cp.match.listener.listener_filter } : {},
          cp.match.listener.name != null ? { name = cp.match.listener.name } : {},
          cp.match.listener.filter_chain != null ? { filterChain = merge(
            cp.match.listener.filter_chain.name != null ? { name = cp.match.listener.filter_chain.name } : {},
            cp.match.listener.filter_chain.sni != null ? { sni = cp.match.listener.filter_chain.sni } : {},
            cp.match.listener.filter_chain.transport_protocol != null ? { transportProtocol = cp.match.listener.filter_chain.transport_protocol } : {},
            cp.match.listener.filter_chain.application_protocols != null ? { applicationProtocols = cp.match.listener.filter_chain.application_protocols } : {},
            cp.match.listener.filter_chain.destination_port != null ? { destinationPort = cp.match.listener.filter_chain.destination_port } : {},
            cp.match.listener.filter_chain.filter != null ? { filter = merge(
              cp.match.listener.filter_chain.filter.name != null ? { name = cp.match.listener.filter_chain.filter.name } : {},
              cp.match.listener.filter_chain.filter.sub_filter != null ? { subFilter = merge(
                cp.match.listener.filter_chain.filter.sub_filter.name != null ? { name = cp.match.listener.filter_chain.filter.sub_filter.name } : {},
              ) } : {},
            ) } : {},
          ) } : {},
        ) } : {},
        cp.match.route_configuration != null ? { routeConfiguration = merge(
          cp.match.route_configuration.port_number != null ? { portNumber = cp.match.route_configuration.port_number } : {},
          cp.match.route_configuration.port_name != null ? { portName = cp.match.route_configuration.port_name } : {},
          cp.match.route_configuration.gateway != null ? { gateway = cp.match.route_configuration.gateway } : {},
          cp.match.route_configuration.name != null ? { name = cp.match.route_configuration.name } : {},
          cp.match.route_configuration.vhost != null ? { vhost = merge(
            cp.match.route_configuration.vhost.name != null ? { name = cp.match.route_configuration.vhost.name } : {},
            cp.match.route_configuration.vhost.domain_name != null ? { domainName = cp.match.route_configuration.vhost.domain_name } : {},
            cp.match.route_configuration.vhost.route != null ? { route = merge(
              cp.match.route_configuration.vhost.route.name != null ? { name = cp.match.route_configuration.vhost.route.name } : {},
              cp.match.route_configuration.vhost.route.action != null ? { action = cp.match.route_configuration.vhost.route.action } : {},
            ) } : {},
          ) } : {},
        ) } : {},
        cp.match.cluster != null ? { cluster = merge(
          cp.match.cluster.port_number != null ? { portNumber = cp.match.cluster.port_number } : {},
          cp.match.cluster.service != null ? { service = cp.match.cluster.service } : {},
          cp.match.cluster.subset != null ? { subset = cp.match.cluster.subset } : {},
          cp.match.cluster.name != null ? { name = cp.match.cluster.name } : {},
        ) } : {},
      ) } : {},
      cp.patch != null ? { patch = merge(
        cp.patch.operation != null ? { operation = cp.patch.operation } : {},
        cp.patch.value != null ? { value = cp.patch.value } : {},
        cp.patch.filter_class != null ? { filterClass = cp.patch.filter_class } : {},
      ) } : {},
    )
  ] : null

  # Assemble the full EnvoyFilter spec, omitting unset optional blocks (and empty lists) so
  # upstream/istiod defaults flow through. An EnvoyFilter with no config_patches is a valid
  # no-op upstream.
  envoy_filter_spec = merge(
    local.workload_selector != null ? { workloadSelector = local.workload_selector } : {},
    local.config_patches != null && length(coalesce(local.config_patches, [])) > 0 ? { configPatches = local.config_patches } : {},
    var.spec.priority != null ? { priority = var.spec.priority } : {},
    local.target_refs != null && length(coalesce(local.target_refs, [])) > 0 ? { targetRefs = local.target_refs } : {},
  )
}
