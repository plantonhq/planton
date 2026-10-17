# Langfuse from the official chart, as a real Helm release (Pulumi twin:
# helm_release.go). The typed spec renders into chart values
# (local.typed_helm_values); the helm_values escape hatch is a SECOND values
# document, merged by the provider with Helm -f semantics; the re-pins are
# merged LAST - the fullname every output derives from, and the four bundled
# stores held off.

resource "helm_release" "langfuse" {
  name       = local.release_name
  repository = local.helm_chart_repo
  chart      = local.helm_chart_name
  version    = local.chart_version
  namespace  = local.namespace

  # The module owns namespace creation (create_namespace flag).
  create_namespace = false

  # The first start runs every Postgres and ClickHouse migration before the
  # web server turns ready, hence the generous timeout.
  wait            = true
  atomic          = true
  cleanup_on_fail = true
  timeout         = 900

  values = concat(
    [yamlencode(local.typed_helm_values)],
    try(var.spec.helm_values, "") != "" ? [var.spec.helm_values] : [],
    [yamlencode(local.re_pinned_values)],
  )

  depends_on = [
    kubernetes_namespace_v1.langfuse,
    kubernetes_secret_v1.instance,
  ]

  lifecycle {
    # NAME BUDGET (Pulumi twin: the length check in main.go Resources): the
    # chart names its worker PodDisruptionBudget <name>-worker-pdb, and an
    # object name is a 63-character DNS label.
    precondition {
      condition     = length(var.metadata.name) <= 52
      error_message = "metadata.name exceeds Langfuse's 52-character name budget (the chart names its worker PodDisruptionBudget <name>-worker-pdb, and an object name is a 63-character DNS label) - shorten metadata.name."
    }
  }
}
