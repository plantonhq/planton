locals {
  service_entry_name = var.metadata.name
  namespace          = var.spec.namespace

  labels = {
    "app.kubernetes.io/name"       = "service-entry"
    "app.kubernetes.io/instance"   = var.metadata.name
    "app.kubernetes.io/managed-by" = "openmcf"
    "app.kubernetes.io/component"  = "service-entry"
  }

  # Ports, proto snake_case mapped to camelCase CRD keys, with unset optional fields
  # pruned per element so they do not reach the manifest.
  ports = var.spec.ports != null ? [
    for port in var.spec.ports : merge(
      { number = port.number, name = port.name },
      port.protocol != null ? { protocol = port.protocol } : {},
      port.target_port != null ? { targetPort = port.target_port } : {},
    )
  ] : null

  # Endpoints, snake_case mapped to camelCase, with unset optional fields pruned per
  # element. The port map and label map pass through unchanged (already string-keyed).
  endpoints = var.spec.endpoints != null ? [
    for endpoint in var.spec.endpoints : merge(
      endpoint.address != null ? { address = endpoint.address } : {},
      endpoint.ports != null ? { ports = endpoint.ports } : {},
      endpoint.labels != null ? { labels = endpoint.labels } : {},
      endpoint.network != null ? { network = endpoint.network } : {},
      endpoint.locality != null ? { locality = endpoint.locality } : {},
      endpoint.weight != null ? { weight = endpoint.weight } : {},
      endpoint.service_account != null ? { serviceAccount = endpoint.service_account } : {},
    )
  ] : null

  # Workload selector, mapped to its camelCase CRD form (labels) and omitted entirely
  # when no labels are provided. The nested labels is read through a ?: conditional
  # (which only evaluates the taken branch) so a null selector never triggers an
  # attribute access on null.
  workload_selector_labels = var.spec.workload_selector != null ? var.spec.workload_selector.labels : null
  workload_selector = local.workload_selector_labels != null ? {
    labels = local.workload_selector_labels
  } : null

  # Assemble the full ServiceEntry spec, omitting unset optional blocks (and empty lists)
  # so upstream/istiod defaults flow through. hosts is always present (required upstream).
  service_entry_spec = merge(
    { hosts = var.spec.hosts },
    var.spec.addresses != null && length(coalesce(var.spec.addresses, [])) > 0 ? { addresses = var.spec.addresses } : {},
    local.ports != null && length(coalesce(local.ports, [])) > 0 ? { ports = local.ports } : {},
    var.spec.location != null ? { location = var.spec.location } : {},
    var.spec.resolution != null ? { resolution = var.spec.resolution } : {},
    local.endpoints != null && length(coalesce(local.endpoints, [])) > 0 ? { endpoints = local.endpoints } : {},
    var.spec.export_to != null && length(coalesce(var.spec.export_to, [])) > 0 ? { exportTo = var.spec.export_to } : {},
    var.spec.subject_alt_names != null && length(coalesce(var.spec.subject_alt_names, [])) > 0 ? { subjectAltNames = var.spec.subject_alt_names } : {},
    local.workload_selector != null ? { workloadSelector = local.workload_selector } : {},
  )
}
