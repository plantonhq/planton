# KubernetesKubePrometheusStack Terraform module.
#
# Installs the kube-prometheus-stack from the official Helm chart as a real
# Helm release. The typed spec renders into chart values
# (locals.helm_values); the CRDs ride the chart's crds subchart
# (install-once, keep-on-uninstall); declared remote-write usernames are
# materialized into a module-owned Secret (the Prometheus CRD reads both
# basic-auth halves from Secrets); typed notification credentials are
# materialized into a second module-owned Secret that Alertmanager mounts;
# the helm_values escape hatch is passed
# as a SECOND values document, which the provider merges over the first
# with Helm -f semantics — the exact semantic twin of the Pulumi module's
# buildHelmValues + mergeMaps.

# The optional installation namespace. Created before the release; deleted
# with the resource.
resource "kubernetes_namespace_v1" "kube_prometheus_stack" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.labels
  }
}

# Module-owned Secret carrying declared remote-write basic-auth usernames
# (key `username-<i>` per entry). Materialized only when an entry declares
# basic auth: the Prometheus CRD reads the username from a Secret, but a
# username is not a secret — the spec accepts it as a plain string and the
# module owns this Secret (the declared-credentials pattern). Passwords
# stay in the user's own Secrets.
resource "kubernetes_secret_v1" "remote_write_auth" {
  count = length(local.remote_write_username_data) > 0 ? 1 : 0

  metadata {
    name      = local.remote_write_auth_secret_name
    namespace = local.namespace
    labels    = local.labels
  }

  data = local.remote_write_username_data

  depends_on = [
    kubernetes_namespace_v1.kube_prometheus_stack,
  ]
}

# Module-owned Secret carrying the typed notification credentials (keys
# `<receiver>-<integration>-<index>-<field>` and `heartbeat-bearer-token`).
# Created BEFORE the release: Alertmanager mounts it through
# alertmanagerSpec.secrets, and a pod whose Secret volume is missing never
# starts. The values arrive resolved from managed-secret references; the
# provider keeps `data` sensitive.
resource "kubernetes_secret_v1" "alertmanager_notifications" {
  count = length(local.notifications_secret_data) > 0 ? 1 : 0

  metadata {
    name      = local.notifications_secret_name
    namespace = local.namespace
    labels    = local.labels
  }

  data = local.notifications_secret_data

  depends_on = [
    kubernetes_namespace_v1.kube_prometheus_stack,
  ]
}

resource "helm_release" "kube_prometheus_stack" {
  name       = local.release_name
  repository = local.helm_chart_repo
  chart      = local.helm_chart_name
  version    = local.chart_version
  namespace  = local.namespace

  # The module owns namespace creation (create_namespace flag).
  create_namespace = false

  # Wait for the whole stack to become Ready — an operator that never
  # starts, an unschedulable Prometheus or an unbindable volume should
  # fail THIS apply, not the first scrape. The budget covers the operator
  # reconciling the Prometheus and Alertmanager StatefulSets after the
  # release's own resources are up.
  wait            = true
  atomic          = true
  cleanup_on_fail = true
  timeout         = 900

  # Documents merged in order by the provider (helm -f semantics): the
  # typed rendering first, the user's escape hatch second — and
  # fullnameOverride re-pinned LAST, the one deliberate exception to the
  # escape hatch's last-word contract (twin of the Pulumi module). Every
  # child service name — and the exported outputs built from them —
  # derives from the fullname; letting an override move it would break
  # every output.
  values = concat(
    [yamlencode(local.helm_values)],
    try(var.spec.helm_values, "") != "" ? [var.spec.helm_values] : [],
    [yamlencode({ fullnameOverride = local.release_name })]
  )

  lifecycle {
    # FAIL LOUDLY on names past the chart's fullname budget: the chart
    # SILENTLY truncates fullnameOverride at 26 characters (headroom for
    # its longest child name), which would break the `<name>-prometheus`
    # / `<name>-grafana` naming contract every exported output is built
    # on. Twin: the Pulumi module's Resources() guard.
    precondition {
      condition     = length(var.metadata.name) <= 26
      error_message = "The kube-prometheus-stack chart derives child names from the resource name and silently truncates past 26 characters, which would break the stack's naming contract — use a name of at most 26 characters."
    }
  }

  depends_on = [
    kubernetes_namespace_v1.kube_prometheus_stack,
    kubernetes_secret_v1.remote_write_auth,
    kubernetes_secret_v1.alertmanager_notifications,
  ]
}
