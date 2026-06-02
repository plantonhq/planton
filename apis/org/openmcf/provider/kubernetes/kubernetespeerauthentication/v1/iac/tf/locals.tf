locals {
  peer_authentication_name = var.metadata.name
  namespace                = var.spec.namespace

  labels = {
    "app.kubernetes.io/name"       = "peer-authentication"
    "app.kubernetes.io/instance"   = var.metadata.name
    "app.kubernetes.io/managed-by" = "openmcf"
    "app.kubernetes.io/component"  = "peer-authentication"
  }

  # Workload selector, mapped to its camelCase CRD form and omitted entirely when
  # no labels are provided so the policy stays namespace/mesh-wide. The nested
  # match_labels is read through a ?: conditional (which only evaluates the taken
  # branch) so a null selector never triggers an attribute access on null.
  selector_match_labels = var.spec.selector != null ? var.spec.selector.match_labels : null
  selector = local.selector_match_labels != null ? {
    matchLabels = local.selector_match_labels
  } : null

  # Per-workload mTLS mode, omitted when unset so the policy inherits from its
  # parent (namespace, then mesh).
  mtls = var.spec.mtls != null ? {
    mode = var.spec.mtls.mode
  } : null

  # Per-port mTLS overrides. The CRD keys this map by port number; Terraform
  # carries the keys as strings, which matches the CRD's string-keyed JSON form.
  port_level_mtls = var.spec.port_level_mtls != null ? {
    for port, port_mtls in var.spec.port_level_mtls : port => {
      mode = port_mtls.mode
    }
  } : null

  # Assemble the full PeerAuthentication spec, omitting unset optional blocks so
  # upstream/controller inheritance behavior flows through.
  peer_authentication_spec = merge(
    local.selector != null ? { selector = local.selector } : {},
    local.mtls != null ? { mtls = local.mtls } : {},
    local.port_level_mtls != null ? { portLevelMtls = local.port_level_mtls } : {},
  )
}
