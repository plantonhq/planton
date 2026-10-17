# Computed values for the KubernetesSolrOperator module. Every resolution
# here has an exact twin in the Pulumi module's locals.go / values.go —
# keep them in lockstep.
#
# HCL DISCIPLINE (applies to every conditional object in this file):
# conditional entries are written as `key = cond ? value : null` inside ONE
# object literal, pruned with `{ for k, v in {...} : k => v if v != null }`.
# The two tempting alternatives are both broken: `cond ? {...} : {}`
# ternaries fail plan-time type unification when branches carry different
# attributes, and `merge(concat(cond ? [{...}] : [], ...)...)` silently
# UNIFIES primitive-only sibling objects into map(string) — numbers and
# booleans arrive in the chart values as strings. The null-prune form
# preserves every value's type.
#
# Optional nested blocks are read with try(): HCL's && does NOT
# short-circuit, so chained null checks still dereference the null — and
# var.spec is typed 'any', so an absent attribute is an error, not a null.

locals {
  # Chart identity — must stay byte-identical with the Pulumi module's
  # vars: cross-engine chart drift deploys two different products from one
  # manifest. Chart versions carry NO `v` prefix (the operator image/CRD
  # artifacts DO — chart 0.9.1 ships operator v0.9.1); the SERVED index
  # (https://solr.apache.org/charts) governs the version.
  helm_chart_repo = "https://solr.apache.org/charts"
  helm_chart_name = "solr-operator"

  # Release name = metadata.name.
  release_name = var.metadata.name

  # Chart version resolved to the pinned default when unset, so both
  # engines install the same chart whether or not the platform's
  # defaulting middleware ran — mirror of the Pulumi module's
  # DefaultChartVersion.
  chart_version = coalesce(try(var.spec.chart_version, null), "0.9.1")

  namespace = var.spec.namespace

  # Resource-identity labels stamped on the namespace this module creates
  # (never injected into the chart's own resources — Helm owns those).
  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesSolrOperator"
    },
    var.metadata.id != null && var.metadata.id != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    var.metadata.org != null && var.metadata.org != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    var.metadata.env != null && var.metadata.env != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  # ---- module-owned CRDs (the derive-branch contract for helm_crds.tf) ----
  # The CRDs are DERIVED from the pinned chart at plan time by the
  # generated helm_crds.tf: it renders the chart with the release's own
  # values (helm_release_values, the same list the release consumes) plus
  # the CRD switch turned on, keeps the CustomResourceDefinition
  # documents, stamps them, and applies each one kept. This object is its
  # input; the twin of the Pulumi module's keptcrds.Args, every key present.
  #
  # The chart carries its CRDs on both of Helm's surfaces: the three
  # solr.apache.org CRDs in its crds/ directory, and the ZookeeperCluster
  # CRD templated by the bundled zookeeper-operator subchart behind
  # zookeeper-operator.crd.create. The render turns that switch on so the
  # module owns all four; the release pins it off (see
  # zookeeper_operator_values). When the subchart is not installed
  # (zookeeper_operator.install = false) its CRD is not rendered either:
  # the set follows the chart's own behaviour at the pin.
  helm_crds_args = {
    # crds.install false is the bring-your-own-CRDs arm (the CRDs are owned
    # elsewhere, a GitOps-managed bundle); the release still skips CRDs.
    # crds.keep_on_uninstall false lets a destroy take the CRDs with it.
    install           = try(var.spec.crds.install, null) == null ? true : var.spec.crds.install
    keep_on_uninstall = try(var.spec.crds.keep_on_uninstall, null) == null ? true : var.spec.crds.keep_on_uninstall

    # A typed kind knows its chart carries CRDs and pins the switch: a render
    # that yields none is a failure, and nothing is ever left to Helm.
    expect_crds = true

    version       = local.chart_version
    version_field = "spec.chart_version"

    chart = {
      repository   = local.helm_chart_repo
      name         = local.helm_chart_name
      release_name = local.release_name
      namespace    = local.namespace
      values       = local.helm_release_values

      # The subchart's CRD switch, turned on for the render only.
      render_override = yamlencode({ "zookeeper-operator" = { crd = { create = true } } })

      # No template in either chart gates on a served API version.
      api_versions       = []
      allow_helm_managed = false

      # Public repository, no set-style overrides.
      repository_username = ""
      repository_password = ""
      set                 = []
      set_sensitive       = []
    }
    bundle = null
  }

  # ---- operator container resources (shared ContainerResources shape) -------
  # Twin of the Pulumi module's resourcesMap. The chart ships NO default
  # resources (`resources: {}` — the operator is lightweight); the key
  # renders only when the spec sets it.
  operator_resources = try(var.spec.resources, null) == null ? null : {
    for k, v in {
      limits = try(var.spec.resources.limits, null) == null ? null : {
        for lk, lv in {
          cpu    = try(var.spec.resources.limits.cpu, "") != "" ? var.spec.resources.limits.cpu : null
          memory = try(var.spec.resources.limits.memory, "") != "" ? var.spec.resources.limits.memory : null
        } : lk => lv if lv != null
      }
      requests = try(var.spec.resources.requests, null) == null ? null : {
        for rk, rv in {
          cpu    = try(var.spec.resources.requests.cpu, "") != "" ? var.spec.resources.requests.cpu : null
          memory = try(var.spec.resources.requests.memory, "") != "" ? var.spec.resources.requests.memory : null
        } : rk => rv if rv != null
      }
    } : k => v if v != null && length(v) > 0
  }

  # ---- bundled zookeeper-operator values -------------------------------------
  # NOTE the dash in the chart's values key: "zookeeper-operator" is the
  # SUBCHART name, quoted wherever it appears as a map key. crd.create
  # is pinned false UNCONDITIONALLY — the module owns the
  # ZookeeperCluster CRD (derived and applied kept by helm_crds.tf), so
  # the subchart must never install its own copy and put it under Helm's
  # delete-on-uninstall lifecycle; the same key is re-pinned after the
  # escape hatch in helm_release_values. install renders on presence
  # (chart default true); use renders only on divergence (chart default
  # false, and it is ignored whenever install is true).
  zookeeper_operator_values = {
    for k, v in {
      crd     = { create = false }
      install = try(var.spec.zookeeper_operator.install, null)
      use     = try(var.spec.zookeeper_operator.use_existing, false) ? true : null
    } : k => v if v != null
  }

  # ---- operator -> Solr mutual TLS --------------------------------------------
  # Secret names arrive as resolved literals (the value-or-ref foreign
  # keys resolve before Terraform runs). Scalars with chart defaults
  # (ca_cert_secret_key, insecure_skip_verify, watch_for_updates) render
  # on presence.
  mtls_values = try(var.spec.mtls, null) == null ? null : {
    for k, v in {
      clientCertSecret   = try(var.spec.mtls.client_cert_secret, "") != "" ? var.spec.mtls.client_cert_secret : null
      caCertSecret       = try(var.spec.mtls.ca_cert_secret, "") != "" ? var.spec.mtls.ca_cert_secret : null
      caCertSecretKey    = try(var.spec.mtls.ca_cert_secret_key, null) != null ? var.spec.mtls.ca_cert_secret_key : null
      insecureSkipVerify = try(var.spec.mtls.insecure_skip_verify, null)
      watchForUpdates    = try(var.spec.mtls.watch_for_updates, null)
    } : k => v if v != null
  }

  # ---- image source ------------------------------------------------------------
  # image.repository / image.tag are the air-gap override; the pull
  # secret is the chart's image.imagePullSecret — a SINGULAR string (the
  # chart accepts exactly one), not the usual imagePullSecrets object
  # list.
  image_values = {
    for k, v in {
      repository      = try(var.spec.image.repository, "") != "" ? var.spec.image.repository : null
      tag             = try(var.spec.image.tag, "") != "" ? var.spec.image.tag : null
      imagePullSecret = try(var.spec.image_pull_secret, "") != "" ? var.spec.image_pull_secret : null
    } : k => v if v != null
  }

  # ---- typed chart values (twin of the Pulumi module's buildHelmValues) ------
  # Chart-default-matching values render only on divergence, so the
  # rendered values stay minimal on both engines. The ONE always-rendered
  # entry is "zookeeper-operator" (it always carries crd.create = false).
  typed_values = {
    for k, v in {
      replicaCount = try(var.spec.replicas, null)
      resources    = local.operator_resources != null && length(local.operator_resources) > 0 ? local.operator_resources : null

      # The chart takes a COMMA-JOINED string (templates/_helpers.tpl
      # splits it back apart), not a YAML list. Empty = the operator
      # watches ALL namespaces (the chart default; also the chart's
      # ClusterRole-vs-Role switch).
      watchNamespaces = length(try(var.spec.watch_namespaces, [])) > 0 ? join(",", var.spec.watch_namespaces) : null

      "zookeeper-operator" = local.zookeeper_operator_values

      # Both nest a single enable flag and default true in the chart.
      # Rendered on presence — an explicit true re-states the chart
      # default harmlessly, an explicit false is the actual opt-out.
      leaderElection = try(var.spec.leader_election_enabled, null) != null ? { enable = var.spec.leader_election_enabled } : null
      metrics        = try(var.spec.metrics_enabled, null) != null ? { enable = var.spec.metrics_enabled } : null

      mTLS = local.mtls_values != null && length(local.mtls_values) > 0 ? local.mtls_values : null

      nodeSelector = length(try(var.spec.node_selector, {})) > 0 ? var.spec.node_selector : null
      tolerations = length(try(var.spec.tolerations, [])) > 0 ? [
        for t in var.spec.tolerations : {
          for tk, tv in {
            key               = try(t.key, "") != "" ? t.key : null
            operator          = try(t.operator, "") != "" ? t.operator : null
            value             = try(t.value, "") != "" ? t.value : null
            effect            = try(t.effect, "") != "" ? t.effect : null
            tolerationSeconds = try(t.toleration_seconds, null)
          } : tk => tv if tv != null
        }
      ] : null

      image = length(local.image_values) > 0 ? local.image_values : null
    } : k => v if v != null
  }

  # ---- deployment name (the chart's fullname template, replayed) -------------
  # templates/_helpers.tpl "solr-operator.fullname" with no
  # nameOverride/fullnameOverride (the typed spec sets neither): the
  # release name itself when it already contains "solr-operator",
  # otherwise "<release>-solr-operator" — truncated to 63 chars with one
  # trailing "-" trimmed. A helm_values override of either template knob
  # would change the real name without being reflected here — the
  # escape-hatch caveat both engines document.
  deployment_name = trimsuffix(
    substr(
      strcontains(local.release_name, local.helm_chart_name) ? local.release_name : "${local.release_name}-${local.helm_chart_name}",
      0, 63
    ),
  "-")

  # ---- the release's values, in helm -f order --------------------------------
  # The typed rendering, the user's escape hatch, then the load-bearing
  # re-pin AFTER the escape hatch so no helm_values override can re-arm
  # it: zookeeper-operator.crd.create=false keeps the subchart's CRD out
  # of Helm's delete-on-uninstall lifecycle (the module owns it). ONE
  # list, consumed by both the release (main.tf) and the CRD render
  # (helm_crds.tf), so the derived CRDs can never see different values
  # than the install.
  helm_release_values = concat(
    [yamlencode(local.typed_values)],
    try(var.spec.helm_values, "") != "" ? [var.spec.helm_values] : [],
    [yamlencode({ "zookeeper-operator" = { crd = { create = false } } })]
  )
}
