# KubernetesPlantonOperator Terraform module.
#
# Installs the Planton operator — the lifecycle manager that reconciles
# PlantonPlatform declarations (KubernetesPlantonPlatform resources) into
# running self-hosted Planton platforms — from the official
# planton-operator Helm chart (OCI, ghcr.io/plantonhq/charts) as ONE real
# Helm release, byte-identical to a hand-installed one.
#
# THE CHART OWNS ITS DEFINITIONS: the PlantonPlatform and
# PlantonIdentityProvider CRDs are release resources the chart renders
# behind its crds.enabled / crds.keep values, so a chart_version upgrade
# carries the matching schema with the operator, and an uninstall keeps
# them by default — destroying the operator never cascade-deletes
# PlantonPlatform declarations or the platforms behind them. This module
# maps the spec's two crds dials onto those values and applies nothing
# else; the optional namespace is the only object it creates itself.
#
# ONE OPERATOR PER CLUSTER: the operator enforces this itself at startup
# (a label-matched Deployment scan that refuses to start beside a
# sibling), so the release name is FIXED — the collision is impossible to
# express instead of merely refused. More PLATFORMS need no second
# operator: one operator watches all namespaces.
#
# The typed spec renders into chart values (locals.typed_values); the
# helm_values escape hatch is passed as a SECOND values document, which
# the provider merges over the first with Helm -f semantics — the exact
# semantic twin of the Pulumi module's buildHelmValues + mergeMaps.

# The optional installation namespace. Created before the release; deleted
# with the resource (pre-existing-namespace installs leave
# create_namespace false).
resource "kubernetes_namespace_v1" "planton_operator" {
  count = try(var.spec.create_namespace, false) ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.labels
  }
}

# The operator release.
resource "helm_release" "planton_operator" {
  name       = local.release_name
  repository = local.chart_repository
  chart      = local.helm_chart_name
  version    = local.chart_version
  namespace  = local.namespace

  # The module owns namespace creation (create_namespace flag). The
  # definitions are release resources the chart renders behind values.crds
  # — never skipped: skip_crds governs only Helm's install-once crds/
  # directory, which this chart does not use.
  create_namespace = false

  # Wait for the operator to become Available — a manager that never
  # becomes ready (the one-per-cluster startup guard refusing beside a
  # sibling operator is THE classic case) should fail THIS apply with a
  # readiness timeout, not surface later as PlantonPlatform resources that
  # mysteriously never reconcile.
  wait            = true
  atomic          = true
  cleanup_on_fail = true
  timeout         = 600

  # Two documents, merged in order by the provider (helm -f semantics):
  # the typed rendering first, the user's escape hatch last.
  values = concat(
    [yamlencode(local.typed_values)],
    try(var.spec.helm_values, "") != "" ? [var.spec.helm_values] : []
  )

  depends_on = [kubernetes_namespace_v1.planton_operator]

  lifecycle {
    # Refuse charts that do not own their definitions: below
    # min_chart_version the crds dials would be silently dropped (older
    # charts have no crds.enabled / crds.keep values), which a module must
    # never do. Twin: the Pulumi module's chartVersionAtLeast guard. HCL
    # has no semver function, so the three parts compare by hand.
    precondition {
      condition = (
        tonumber(split(".", local.chart_version)[0]) > tonumber(split(".", local.min_chart_version)[0])
        ) || (
        tonumber(split(".", local.chart_version)[0]) == tonumber(split(".", local.min_chart_version)[0])
        && tonumber(split(".", local.chart_version)[1]) > tonumber(split(".", local.min_chart_version)[1])
        ) || (
        tonumber(split(".", local.chart_version)[0]) == tonumber(split(".", local.min_chart_version)[0])
        && tonumber(split(".", local.chart_version)[1]) == tonumber(split(".", local.min_chart_version)[1])
        && tonumber(split(".", local.chart_version)[2]) >= tonumber(split(".", local.min_chart_version)[2])
      )
      error_message = "observed: spec.chart_version is ${local.chart_version}\nmeaning: planton-operator charts older than ${local.min_chart_version} install their definitions once from Helm's crds/ directory and have no crds.enabled / crds.keep values, so the crds dials of this resource would have no effect\nnext step: set spec.chart_version to ${local.min_chart_version} or newer (or leave it unset for the catalog default)"
    }
  }
}
