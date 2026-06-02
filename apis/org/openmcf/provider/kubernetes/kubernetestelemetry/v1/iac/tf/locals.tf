locals {
  telemetry_name = var.metadata.name
  namespace      = var.spec.namespace

  labels = {
    "app.kubernetes.io/name"       = "telemetry"
    "app.kubernetes.io/instance"   = var.metadata.name
    "app.kubernetes.io/managed-by" = "openmcf"
    "app.kubernetes.io/component"  = "telemetry"
  }

  # ----------------------------------------------------------------------------------------
  # Null-pruning shape rules (so unset fields are omitted and istiod defaults flow through):
  #
  #  - OBJECT-typed oneOf members -> merge() of conditional fragments (omit the non-chosen
  #    members entirely). custom_tags' {literal|environment|header} are objects with required
  #    subfields; emitting a non-chosen member as null would be sent as an empty object {},
  #    which both violates its required subfields and matches a second oneOf arm. So only the
  #    chosen member is emitted, with its required subfields seeded as the merge base.
  #
  #  - SCALAR oneOf members and uniform-type leaves -> explicit object constructor with
  #    value-or-null (NOT a conditional merge). An all-conditional merge() of fields that are
  #    all the same scalar/collection type collapses to a map() that kubernetes_manifest
  #    cannot morph into an object ("Map element into Object"). The metrics override `match`
  #    (metric/custom_metric/mode -- all strings, with a metric-vs-custom_metric oneOf) and
  #    tag_overrides values ({operation, value} -- both strings) are built this way;
  #    kubernetes_manifest prunes the SCALAR nulls, so the metric/custom_metric oneOf still
  #    sees only the field that was set.
  # ----------------------------------------------------------------------------------------

  # Workload selector, mapped to its camelCase CRD form and omitted when no labels are
  # provided. match_labels is read through a ?: conditional (only the taken branch is
  # evaluated) so a null selector never triggers an attribute access on null.
  selector_match_labels = var.spec.selector != null ? var.spec.selector.match_labels : null
  selector              = local.selector_match_labels != null ? { matchLabels = local.selector_match_labels } : null

  # Target references, proto snake_case mapped to camelCase, with unset optional fields
  # pruned per element.
  target_refs = var.spec.target_refs != null ? [
    for ref in var.spec.target_refs : merge(
      { kind = ref.kind, name = ref.name },
      ref.group != null ? { group = ref.group } : {},
      ref.namespace != null ? { namespace = ref.namespace } : {},
    )
  ] : null

  # Tracing rules. Mixed-type fields, so each rule is merge()-pruned. custom_tags values are
  # merge()-pruned (object-typed oneOf). The single-field `match` selector is an object
  # constructor ({mode}).
  tracing = var.spec.tracing != null ? [
    for t in var.spec.tracing : merge(
      t.match != null ? { match = { mode = t.match.mode } } : {},
      t.providers != null ? { providers = [for p in t.providers : { name = p.name }] } : {},
      t.random_sampling_percentage != null ? { randomSamplingPercentage = t.random_sampling_percentage } : {},
      t.disable_span_reporting != null ? { disableSpanReporting = t.disable_span_reporting } : {},
      t.custom_tags != null ? { customTags = {
        for k, tag in t.custom_tags : k => merge(
          tag.literal != null ? { literal = { value = tag.literal.value } } : {},
          tag.environment != null ? { environment = merge(
            { name = tag.environment.name },
            tag.environment.default_value != null ? { defaultValue = tag.environment.default_value } : {},
          ) } : {},
          tag.header != null ? { header = merge(
            { name = tag.header.name },
            tag.header.default_value != null ? { defaultValue = tag.header.default_value } : {},
          ) } : {},
        )
      } } : {},
      t.enable_istio_tags != null ? { enableIstioTags = t.enable_istio_tags } : {},
      t.use_request_id_for_trace_sampling != null ? { useRequestIdForTraceSampling = t.use_request_id_for_trace_sampling } : {},
    )
  ] : null

  # Metrics rules. The override `match` is an object constructor (scalar metric-vs-custom_metric
  # oneOf + mode), and tag_overrides values are object constructors ({operation, value}); both
  # rely on kubernetes_manifest pruning scalar nulls.
  metrics = var.spec.metrics != null ? [
    for m in var.spec.metrics : merge(
      m.providers != null ? { providers = [for p in m.providers : { name = p.name }] } : {},
      m.overrides != null ? { overrides = [
        for o in m.overrides : merge(
          o.match != null ? { match = {
            metric       = o.match.metric
            customMetric = o.match.custom_metric
            mode         = o.match.mode
          } } : {},
          o.disabled != null ? { disabled = o.disabled } : {},
          o.tag_overrides != null ? { tagOverrides = {
            for k, tag in o.tag_overrides : k => {
              operation = tag.operation
              value     = tag.value
            }
          } } : {},
        )
      ] } : {},
      m.reporting_interval != null ? { reportingInterval = m.reporting_interval } : {},
    )
  ] : null

  # Access logging rules. Mixed-type fields, so merge()-pruned; the single-field `match`
  # selector and `filter` are object constructors.
  access_logging = var.spec.access_logging != null ? [
    for al in var.spec.access_logging : merge(
      al.match != null ? { match = { mode = al.match.mode } } : {},
      al.providers != null ? { providers = [for p in al.providers : { name = p.name }] } : {},
      al.disabled != null ? { disabled = al.disabled } : {},
      al.filter != null ? { filter = { expression = al.filter.expression } } : {},
    )
  ] : null

  # Assemble the full Telemetry spec, omitting unset optional blocks (and empty lists) so
  # upstream/controller behavior flows through.
  telemetry_spec = merge(
    local.selector != null ? { selector = local.selector } : {},
    local.target_refs != null && length(coalesce(local.target_refs, [])) > 0 ? { targetRefs = local.target_refs } : {},
    local.tracing != null && length(coalesce(local.tracing, [])) > 0 ? { tracing = local.tracing } : {},
    local.metrics != null && length(coalesce(local.metrics, [])) > 0 ? { metrics = local.metrics } : {},
    local.access_logging != null && length(coalesce(local.access_logging, [])) > 0 ? { accessLogging = local.access_logging } : {},
  )
}
