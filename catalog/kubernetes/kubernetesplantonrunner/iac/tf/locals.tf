# Computed values for the KubernetesPlantonRunner module. Every resolution
# here has an exact twin in the Pulumi module — keep them in lockstep: same
# rendered chart values, same materialized Secret, same outputs.

locals {
  # Pinned OCI registry path and chart name; chart_version resolves to the
  # pinned default when unset — the version this catalog release was
  # validated against (0.5.0: the pod is probed over gRPC on the runner's
  # port, the contract every execution mode serves; 0.4.0 probed the tunnel
  # agent's HTTP port and restarted a runner enrolled with a tunnel-less
  # instance every liveness window). min_chart_version is the
  # enrollment-contract floor: charts below 0.4.0 predate token enrollment
  # and silently IGNORE the enrollment values (main.tf's precondition refuses
  # them loudly).
  helm_oci_repo         = "oci://ghcr.io/plantonhq/charts"
  helm_chart_name       = "planton-runner"
  default_chart_version = "0.5.0"
  min_chart_version     = "0.4.0"
  chart_version         = try(var.spec.chart_version, "") != "" ? var.spec.chart_version : local.default_chart_version

  namespace    = var.spec.namespace
  release_name = var.metadata.name

  # The name the runner registers itself under when it joins the control
  # plane. spec.runner_name, falling back to "<env>-<metadata.name>"
  # (metadata.name outside an environment) — the SAME derivation the
  # platform uses for records that reference this runner (its minted
  # token, its managed destroy); changing this formula breaks arrival
  # attribution and managed teardown. Pulumi twin: locals.RunnerName.
  runner_name = try(var.spec.runner_name, "") != "" ? var.spec.runner_name : (
    try(var.metadata.env, "") != "" ? "${var.metadata.env}-${var.metadata.name}" : var.metadata.name
  )

  # The module-created Secret the chart reads the runner token from (its
  # existingSecret form) — the token never rides rendered chart values.
  token_secret_name = "${var.metadata.name}-token"
  token_secret_key  = "token"

  # Planton governance labels for the module-created satellites (the
  # namespace and the token Secret — never injected into the chart's own
  # resources; Helm owns those).
  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesPlantonRunner"
    },
    try(var.metadata.id, "") != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    try(var.metadata.org, "") != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    try(var.metadata.env, "") != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  # ---- enrollment (the chart's token contract) ---------------------------------
  # Always the existingSecret form — a Secret NAME, never the token
  # itself (rendered values land in Helm's release Secret, where an
  # inline token would be readable by anyone with release-history read
  # access). The endpoint is contributed only when set: the runner's
  # built-in hosted default applies otherwise, and the chart then renders
  # no PLANTON_RUNNER_ENDPOINT env at all.
  enrollment_block = merge(
    {
      existingSecret    = local.token_secret_name
      existingSecretKey = local.token_secret_key
      runnerName        = local.runner_name
    },
    try(var.spec.control_plane_endpoint, "") != "" ? { endpoint = var.spec.control_plane_endpoint } : {}
  )

  # ---- container sizing ----------------------------------------------------------
  # Rendered ONLY when customized: the chart's own defaults (requests
  # 100m/256Mi, limits 1/1Gi) are the documented baseline, and an empty
  # requests/limits map would REPLACE them with nothing.
  resources_block = try(var.spec.resources, null) == null ? null : {
    for k, v in {
      requests = try(var.spec.resources.requests, null) == null ? null : {
        for rk, rv in {
          cpu    = var.spec.resources.requests.cpu
          memory = var.spec.resources.requests.memory
        } : rk => rv if rv != null && rv != ""
      }
      limits = try(var.spec.resources.limits, null) == null ? null : {
        for rk, rv in {
          cpu    = var.spec.resources.limits.cpu
          memory = var.spec.resources.limits.memory
        } : rk => rv if rv != null && rv != ""
      }
    } : k => v if v != null
  }

  # ---- build worker -----------------------------------------------------------------
  build_block = try(var.spec.build, null) == null || !try(var.spec.build.enabled, false) ? null : merge(
    { enabled = true },
    try(var.spec.build.tekton_namespace, "") != "" ? { tektonNamespace = var.spec.build.tekton_namespace } : {}
  )

  # ---- typed chart values (Pulumi twin: buildHelmValues) ------------------------------
  # fullnameOverride pins the chart's child names to the resource name: the
  # default helper would name every child `<name>-planton-runner`, which
  # wastes name budget and decouples child names from the resource.
  typed_helm_values = merge(
    {
      fullnameOverride = local.release_name
      image = {
        repository = var.spec.image_repository
        tag        = var.spec.runner_version
      }
      enrollment = local.enrollment_block
    },
    local.resources_block != null && local.resources_block != {} ? { resources = local.resources_block } : {},
    local.build_block != null ? { build = local.build_block } : {}
  )
}
