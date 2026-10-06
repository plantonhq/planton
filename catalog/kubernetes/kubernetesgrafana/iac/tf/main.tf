# KubernetesGrafana Terraform module.
#
# Installs Grafana from the official Helm chart as a real Helm release. The
# typed spec renders into chart values (locals.helm_values); admin
# credentials stay chart-owned (generated once via the chart's lookup, or
# read from an existing Secret); database, datasource and sign-in
# credentials ride environment variables sourced from Secrets (sign-in's from
# the module-owned `<name>-sso` Secret below), so no credential ever lands in
# the chart's rendered configuration; the helm_values escape hatch is
# passed as a SECOND values document, which the provider merges over the
# first with Helm -f semantics — the exact semantic twin of the Pulumi
# module's buildHelmValues + mergeMaps.

# The optional installation namespace. Created before the release; deleted
# with the resource.
resource "kubernetes_namespace_v1" "grafana" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.labels
  }
}

# Module-owned Secret carrying the sign-in client secrets (keys
# `google-client-secret`, `generic-oauth-client-secret`). Created BEFORE the
# release: Grafana's GF_AUTH_*_CLIENT_SECRET variables read it through
# secretKeyRef, and a pod whose referenced key is missing never starts. The
# values arrive resolved from managed-secret references; the provider keeps
# `data` sensitive.
resource "kubernetes_secret_v1" "sso" {
  count = local.sign_in_enabled ? 1 : 0

  metadata {
    name      = local.sso_secret_name
    namespace = local.namespace
    labels    = local.labels
  }

  data = local.sso_secret_data

  depends_on = [
    kubernetes_namespace_v1.grafana,
  ]
}

resource "helm_release" "grafana" {
  name       = local.release_name
  repository = local.helm_chart_repo
  chart      = local.helm_chart_name
  version    = local.chart_version
  namespace  = local.namespace

  # The module owns namespace creation (create_namespace flag).
  create_namespace = false

  # Wait for Grafana to become Ready — a UI that never starts (bad plugin
  # ID, unreachable database, unbindable volume) should fail THIS apply,
  # not the first login attempt. Plugin downloads at startup are what the
  # generous budget covers.
  wait            = true
  atomic          = true
  cleanup_on_fail = true
  timeout         = 600

  # Documents merged in order by the provider (helm -f semantics): the
  # typed rendering first, the user's escape hatch second — and
  # fullnameOverride re-pinned LAST, the one deliberate exception to the
  # escape hatch's last-word contract (twin of the Pulumi module). The
  # Service and the chart-generated admin Secret — and the exported
  # outputs built from them — derive from the fullname; letting an
  # override move it would break every output.
  values = concat(
    [yamlencode(local.helm_values)],
    try(var.spec.helm_values, "") != "" ? [var.spec.helm_values] : [],
    [yamlencode({ fullnameOverride = local.release_name })]
  )

  lifecycle {
    # NAME BUDGET: the agent reader's Job is `<name>-agent-reader-<8 hex>`
    # and Kubernetes caps a Job's name at 63 characters. Refused before
    # anything is created (twin of the check in the Pulumi module's
    # main.go).
    precondition {
      condition     = !local.agent_reader_declared || length(local.release_name) <= 41
      error_message = "metadata.name is longer than 41 characters, the most agent_reader allows: the agent reader's Job is <name>-agent-reader-<8 hex> under Kubernetes' 63-character Job name cap. Use a shorter name."
    }
  }

  depends_on = [
    kubernetes_namespace_v1.grafana,
    kubernetes_secret_v1.sso,
  ]
}
