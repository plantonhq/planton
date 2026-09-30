# KubernetesLoki Terraform module.
#
# Installs Grafana Loki from the official Helm chart as a real Helm release.
# The typed spec renders into chart values (locals.helm_values); exactly one
# deployment mode renders with every other mode's workloads zeroed; declared
# object-store credentials ride environment variables sourced from Secrets
# so no credential ever lands in the chart's rendered configuration; the
# helm_values escape hatch is passed as a SECOND values document, with
# fullnameOverride re-pinned in a THIRD — the exact semantic twin of the
# Pulumi module's buildHelmValues + mergeMaps + re-pin.

# The optional installation namespace. Created before the release; deleted
# with the resource.
resource "kubernetes_namespace_v1" "loki" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.labels
  }
}

# Module-owned Secret carrying the r2 arm's key pair (keys `access-key-id`,
# `secret-access-key`). Created BEFORE the release: Loki's credential
# variables read it through secretKeyRef, and a pod whose referenced key is
# missing never starts. The values arrive resolved from references; the
# provider keeps `data` sensitive.
resource "kubernetes_secret_v1" "r2_credentials" {
  count = local.r2 != null ? 1 : 0

  metadata {
    name      = local.r2_secret_name
    namespace = local.namespace
    labels    = local.labels
  }

  data = local.r2_secret_data

  depends_on = [
    kubernetes_namespace_v1.loki,
  ]
}

resource "helm_release" "loki" {
  name       = local.release_name
  repository = local.helm_chart_repo
  chart      = local.helm_chart_name
  version    = local.chart_version
  namespace  = local.namespace

  # The module owns namespace creation (create_namespace flag).
  create_namespace = false

  # Wait for Loki to become Ready — a log store whose ingesters never bind
  # their storage or whose gateway never starts should fail THIS apply, not
  # the first push.
  wait            = true
  atomic          = true
  cleanup_on_fail = true
  timeout         = 600

  # Documents merged in order by the provider (helm -f semantics): the typed
  # rendering first, the user's escape hatch second, and fullnameOverride
  # re-pinned LAST — the one deliberate exception to the escape hatch's
  # last-word contract (twin of the Pulumi module). Every child name — and
  # the exported outputs built from them — derives from the fullname;
  # letting an override move it would break every output.
  values = concat(
    [yamlencode(local.helm_values)],
    try(var.spec.helm_values, "") != "" ? [var.spec.helm_values] : [],
    [yamlencode({ fullnameOverride = local.release_name })]
  )

  # The chart composes child names like `<name>-backend-headless` and
  # truncates the COMPOSED name at 63 characters — an over-long resource
  # name corrupts the naming contract the outputs promise. Fail THIS plan
  # loudly instead (twin: the Pulumi module's MaxNameLength guard).
  lifecycle {
    precondition {
      condition     = local.name_within_budget
      error_message = "metadata.name '${local.release_name}' is ${length(local.release_name)} characters — the loki chart's child-name budget allows at most ${local.max_name_length} (it composes names like <name>-backend-headless within Kubernetes' 63-character cap)."
    }
  }

  depends_on = [
    kubernetes_namespace_v1.loki,
    kubernetes_secret_v1.r2_credentials,
  ]
}
