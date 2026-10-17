# Computed values shared across the module.

locals {
  # Resource-identity labels: the kuberneteslabelkeys set, identical to what
  # the Pulumi module stamps for the same manifest. Stamped on the namespace
  # this module creates — never injected into the chart's own resources
  # (Helm owns those).
  identity_labels_base = {
    "planton.ai/resource" = "true"
    "planton.ai/name"     = var.metadata.name
    "planton.ai/kind"     = "KubernetesHelmRelease"
  }

  id_label = (
    var.metadata.id != null && try(var.metadata.id, "") != ""
  ) ? { "planton.ai/id" = var.metadata.id } : {}

  org_label = (
    var.metadata.org != null && try(var.metadata.org, "") != ""
  ) ? { "planton.ai/organization" = var.metadata.org } : {}

  env_label = (
    var.metadata.env != null && try(var.metadata.env, "") != ""
  ) ? { "planton.ai/environment" = var.metadata.env } : {}

  labels = merge(
    local.identity_labels_base,
    local.id_label,
    local.org_label,
    local.env_label,
  )

  # The namespace the release installs into (the generator flattens the
  # spec's value-or-ref to a plain string) and whether this module creates
  # it.
  namespace_name   = var.spec.namespace
  create_namespace = try(var.spec.create_namespace, false)

  # The Helm release name: spec.release_name when set, otherwise the
  # resource's metadata.name — identical resolution in the Pulumi module.
  release_name = try(var.spec.release_name, "") != "" ? var.spec.release_name : var.metadata.name

  # Helm's own defaults resolved for the optional knobs so both engines send
  # identical values whether or not the spec set the fields (the Pulumi
  # module resolves the same defaults in its locals).
  timeout_seconds = try(var.spec.timeout_seconds, null) != null ? var.spec.timeout_seconds : 300
  max_history     = try(var.spec.max_history, null) != null ? var.spec.max_history : 10

  # skip_await inverts helm_release's wait flag (the Pulumi module passes
  # SkipAwait directly). Both engines default to awaiting readiness.
  wait = !try(var.spec.skip_await, false)

  # ---- chart identity, in the names the shared helm_crds.tf reads ----------
  # Every Helm kind carries these locals; here they come from the spec
  # rather than module constants, because the chart is the user's.
  helm_chart_repo = var.spec.repo
  helm_chart_name = var.spec.chart
  chart_version   = var.spec.version
  namespace       = local.namespace_name

  # ---- the release's values, in helm -f order --------------------------------
  # values_yaml is the one values document; the set-style layers ride the
  # release's set/set_sensitive attributes and are handed to the CRD render
  # the same way (helm_crds_args below), so the render can never see
  # different values than the install.
  helm_release_values = try(var.spec.values_yaml, "") != "" ? [var.spec.values_yaml] : []

  helm_set = concat(
    [for k, v in try(var.spec.set, {}) : { name = k, value = v, type = "auto" }],
    [for k, v in try(var.spec.set_string, {}) : { name = k, value = v, type = "string" }],
  )
  helm_set_sensitive = [
    for k, v in try(var.spec.set_sensitive, {}) : { name = k, value = v, type = "string" }
  ]

  # ---- module-owned CRDs (the derive-branch contract for helm_crds.tf) ----
  # The CRDs are DERIVED from the pinned chart at plan time by the
  # generated helm_crds.tf, rendered with exactly the values the release
  # installs with. This object is its input; the twin of the Pulumi
  # module's keptcrds.Args, every key present.
  #
  # The chart is arbitrary, so the module supplies no render override and
  # ownership follows Helm's two surfaces: the chart's crds/ directory is
  # the module's (derived, kept, moved with the version, re-adopted,
  # never downgraded); CRDs the chart templates stay Helm's, refused
  # unless the chart keeps them itself or allow_helm_managed accepts them.
  # A chart without CRDs (most charts) is ordinary: nothing is applied.
  helm_crds_args = {
    install           = try(var.spec.crds.install, null) == null ? true : var.spec.crds.install
    keep_on_uninstall = try(var.spec.crds.keep_on_uninstall, null) == null ? true : var.spec.crds.keep_on_uninstall
    expect_crds       = false

    version       = local.chart_version
    version_field = "spec.version"

    chart = {
      repository   = local.helm_chart_repo
      name         = local.helm_chart_name
      release_name = local.release_name
      namespace    = local.namespace
      values       = local.helm_release_values

      render_override    = ""
      api_versions       = []
      allow_helm_managed = try(var.spec.crds.allow_helm_managed, null) == null ? false : var.spec.crds.allow_helm_managed

      # The same credentials and set layers the release uses.
      repository_username = try(var.spec.repository_username, "")
      repository_password = try(var.spec.repository_password, "")
      set                 = local.helm_set
      set_sensitive       = local.helm_set_sensitive
    }
    bundle = null
  }
}
