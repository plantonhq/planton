locals {
  authorization_policy_name = var.metadata.name
  namespace                 = var.spec.namespace

  labels = {
    "app.kubernetes.io/name"       = "authorization-policy"
    "app.kubernetes.io/instance"   = var.metadata.name
    "app.kubernetes.io/managed-by" = "openmcf"
    "app.kubernetes.io/component"  = "authorization-policy"
  }

  # Workload selector, mapped to its camelCase CRD form and omitted entirely when
  # no labels are provided. The nested match_labels is read through a ?: conditional
  # (which only evaluates the taken branch) so a null selector never triggers an
  # attribute access on null.
  selector_match_labels = var.spec.selector != null ? var.spec.selector.match_labels : null
  selector = local.selector_match_labels != null ? {
    matchLabels = local.selector_match_labels
  } : null

  # Target references, proto snake_case mapped to camelCase, with unset optional
  # fields pruned per element so they do not reach the manifest.
  target_refs = var.spec.target_refs != null ? [
    for ref in var.spec.target_refs : merge(
      { kind = ref.kind, name = ref.name },
      ref.group != null ? { group = ref.group } : {},
      ref.namespace != null ? { namespace = ref.namespace } : {},
    )
  ] : null

  # Rules. Each rule's from/to/when blocks are null-pruned at the rule level (a
  # rule with none set collapses to {}, the upstream allow-all/deny-all rule). The
  # `source` and `operation` leaves are built as explicit object constructors with
  # value-or-null per field (snake_case -> camelCase) rather than the merge-pruning
  # used elsewhere: their attributes are ALL the same type (list(string)) and all
  # optional, and an all-conditional uniform-type `merge()` collapses to a
  # map(list(string)) in HCL, which `kubernetes_manifest` rejects ("Failed to
  # transform Map element into Object element type"). An object constructor always
  # yields an object type; kubernetes_manifest treats the null attributes as unset.
  # (Components like RequestAuthentication's target_refs avoid the collapse via a
  # fixed-key base; HTTPRoute's blocks avoid it via mixed value types.)
  rules = var.spec.rules != null ? [
    for r in var.spec.rules : merge(
      r.from != null ? { from = [
        for f in r.from : merge(
          f.source != null ? { source = {
            principals           = f.source.principals
            notPrincipals        = f.source.not_principals
            requestPrincipals    = f.source.request_principals
            notRequestPrincipals = f.source.not_request_principals
            namespaces           = f.source.namespaces
            notNamespaces        = f.source.not_namespaces
            serviceAccounts      = f.source.service_accounts
            notServiceAccounts   = f.source.not_service_accounts
            ipBlocks             = f.source.ip_blocks
            notIpBlocks          = f.source.not_ip_blocks
            remoteIpBlocks       = f.source.remote_ip_blocks
            notRemoteIpBlocks    = f.source.not_remote_ip_blocks
          } } : {}
        )
      ] } : {},
      r.to != null ? { to = [
        for t in r.to : merge(
          t.operation != null ? { operation = {
            hosts      = t.operation.hosts
            notHosts   = t.operation.not_hosts
            ports      = t.operation.ports
            notPorts   = t.operation.not_ports
            methods    = t.operation.methods
            notMethods = t.operation.not_methods
            paths      = t.operation.paths
            notPaths   = t.operation.not_paths
          } } : {}
        )
      ] } : {},
      r.when != null ? { when = [
        for c in r.when : {
          key       = c.key
          values    = c.values
          notValues = c.not_values
        }
      ] } : {},
    )
  ] : null

  # Assemble the full AuthorizationPolicy spec, omitting unset optional blocks (and
  # empty lists) so upstream/controller behavior flows through. `action` defaults to
  # ALLOW upstream when omitted; `provider` is only meaningful with the CUSTOM action.
  authorization_policy_spec = merge(
    local.selector != null ? { selector = local.selector } : {},
    local.target_refs != null && length(coalesce(local.target_refs, [])) > 0 ? { targetRefs = local.target_refs } : {},
    local.rules != null && length(coalesce(local.rules, [])) > 0 ? { rules = local.rules } : {},
    var.spec.action != null ? { action = var.spec.action } : {},
    var.spec.provider != null ? { provider = { name = var.spec.provider.name } } : {},
  )
}
