# KubernetesTempo Terraform module.
#
# Installs Grafana Tempo from the official monolithic Helm chart as a real
# Helm release. The typed spec renders into chart values (locals.helm_values);
# declared object-store credentials ride environment variables sourced from
# Secrets so no credential ever lands in the chart's rendered configuration;
# the helm_values escape hatch is a SECOND values document with
# fullnameOverride re-pinned in a THIRD — the exact semantic twin of the
# Pulumi module.

resource "kubernetes_namespace_v1" "tempo" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.labels
  }
}

# Module-owned Secret carrying the r2 arm's key pair (keys `access-key-id`,
# `secret-access-key`). Created BEFORE the release: Tempo's credential
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
    kubernetes_namespace_v1.tempo,
  ]
}

resource "helm_release" "tempo" {
  name       = local.release_name
  repository = local.helm_chart_repo
  chart      = local.helm_chart_name
  version    = local.chart_version
  namespace  = local.namespace

  create_namespace = false

  wait            = true
  atomic          = true
  cleanup_on_fail = true
  timeout         = 600

  values = concat(
    [yamlencode(local.helm_values)],
    try(var.spec.helm_values, "") != "" ? [var.spec.helm_values] : [],
    [yamlencode({ fullnameOverride = local.release_name })]
  )

  lifecycle {
    precondition {
      condition     = local.name_within_budget
      error_message = "metadata.name '${local.release_name}' is ${length(local.release_name)} characters — the tempo chart's child-name budget allows at most ${local.max_name_length} (composed names must stay within Kubernetes' 63-character cap)."
    }
  }

  depends_on = [
    kubernetes_namespace_v1.tempo,
    kubernetes_secret_v1.r2_credentials,
  ]
}
