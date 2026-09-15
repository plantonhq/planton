# KubernetesOpenBao Terraform module.
#
# Installs OpenBao from the official chart as a real Helm release. The
# typed spec renders into chart values AND into the server's HCL
# configuration (locals.bao_config_hcl) — the chart takes config as a raw
# string, so the module owns synthesizing the listener, storage, Raft
# retry_join, seal and telemetry stanzas from typed fields. Exact twin of
# the Pulumi module (bao_config.go / values.go).
#
# THE SEAL LIFECYCLE (why this module never waits on the release): a
# fresh OpenBao server is UNINITIALIZED and SEALED, and the chart's
# readiness probe (`bao status`) deliberately fails for sealed servers —
# pod readiness IS the seal status. `bao operator init` + unseal are
# runtime API operations no deployment tool performs; until someone
# performs them the StatefulSet never reports ready, so a Helm wait here
# would hang every fresh install. The chart keeps sealed pods addressable
# (publishNotReadyAddresses on every Service) exactly so init/unseal can
# reach them.

# The optional installation namespace. Created before the release;
# deleted with the resource.
resource "kubernetes_namespace_v1" "openbao" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.labels
  }
}

# Declared-credential auto-unseal arms materialize their material into a
# module-owned Secret BEFORE the release; the chart wires each key to the
# server as an environment variable (extraSecretEnvironmentVars) — the
# config ConfigMap carries only non-credential seal parameters.
resource "kubernetes_secret_v1" "seal_credentials" {
  count = length(local.seal_secret_data) > 0 ? 1 : 0

  metadata {
    name      = local.seal_credentials_secret_name
    namespace = local.namespace
    labels    = local.labels
  }

  data = local.seal_secret_data

  depends_on = [
    kubernetes_namespace_v1.openbao,
  ]
}

resource "helm_release" "openbao" {
  name       = local.release_name
  repository = local.helm_chart_repo
  chart      = local.helm_chart_name
  version    = local.chart_version
  namespace  = local.namespace

  # The module owns namespace creation (create_namespace flag).
  create_namespace = false

  # NEVER wait, NEVER atomic: sealed/uninitialized pods are NotReady BY
  # DESIGN (see the module header) — a wait hangs every fresh install
  # and atomic would then roll it back. The Pulumi twin sets
  # SkipAwait: true / Atomic: false for the same reason. The E2E
  # verifier owns initialization and readiness.
  wait            = false
  wait_for_jobs   = false
  atomic          = false
  cleanup_on_fail = false
  timeout         = 600

  # Documents merged in order by the provider (helm -f semantics): the
  # typed rendering first, the user's escape hatch second — and
  # fullnameOverride re-pinned LAST, the one deliberate exception to the
  # escape hatch's last-word contract (twin of the Pulumi module). Every
  # exported name output and every synthesized retry_join address
  # derives from the fullname; letting an override move it would break
  # the Raft cluster's own peer list.
  values = concat(
    [yamlencode(local.typed_helm_values)],
    try(var.spec.helm_values, "") != "" ? [var.spec.helm_values] : [],
    [yamlencode({ fullnameOverride = local.release_name })]
  )

  lifecycle {
    # NAME BUDGET (chart truth at 0.28.6): the chart truncates its
    # fullname at 63 then APPENDS Service suffixes — `-internal` (9)
    # always, `-agent-injector-svc` (19) with the injector — and
    # Kubernetes caps Service names at 63. The Pulumi twin enforces the
    # same budget.
    precondition {
      condition     = length(var.metadata.name) <= local.max_name_length
      error_message = "metadata.name exceeds the OpenBao name budget: the chart derives Service names by suffixing (up to 19 characters with the injector enabled) and the backup CronJob is <name>-backup under Kubernetes' 52-character CronJob cap, so use at most 44 characters with the injector, 45 with backup declared, or 54 otherwise."
    }
  }

  depends_on = [
    kubernetes_namespace_v1.openbao,
    kubernetes_secret_v1.seal_credentials,
  ]
}
