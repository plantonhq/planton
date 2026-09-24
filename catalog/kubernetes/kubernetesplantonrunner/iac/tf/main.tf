# KubernetesPlantonRunner Terraform module.
#
# Deploys a standing Planton runner from the official planton-runner OCI
# chart as a real Helm release, so the deployed runner is byte-identical
# to a hand-installed one — the chart carries the load-bearing enrollment
# mechanics this module deliberately does NOT re-model: replicas pinned to
# 1 with a Recreate strategy (two live pods under one runner name would
# revoke each other's keys), the ephemeral identity volume the runner
# persists its minted identity into (container restarts reuse it; pod
# recreation re-joins with the token), and the health endpoints.
#
# SECRET DISCIPLINE: the module materializes the runner token as the
# `<name>-token` Secret BEFORE the release and passes only its NAME into
# chart values (the chart's existingSecret form) — token material never
# rides rendered values, and the escape hatch cannot move it (the
# enrollment block is re-pinned as the last values document).
#
# NO HELM WAIT: the runner's readiness contract is its control-plane work
# queue, not pod liveness — a runner whose control plane is momentarily
# unreachable (or whose token is still propagating) must still deploy and
# destroy cleanly. Helm --wait would couple the install to control-plane
# reachability; the E2E verifier owns the install-level proof instead
# (the same posture as the runner appliances on the cloud substrates).

# The optional installation namespace. Created before the release; deleted
# with the resource.
resource "kubernetes_namespace_v1" "planton_runner" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.labels
  }
}

# The materialized runner token — the chart's existingSecret contract.
# The token authorizes JOINING and is never the runner's identity: the
# runner presents it once per enrollment, registers itself, and persists
# the identity it receives on its own volume. Rotating the token updates
# this Secret; running pods keep serving on their minted identity (the
# token is only read at join), and the next pod recreation joins with the
# new token. Pulumi twin: tokenSecret.
resource "kubernetes_secret_v1" "runner_token" {
  metadata {
    name      = local.token_secret_name
    namespace = local.namespace
    labels    = local.labels
  }

  type = "Opaque"

  data = {
    (local.token_secret_key) = var.spec.token
  }

  depends_on = [
    kubernetes_namespace_v1.planton_runner,
  ]
}

resource "helm_release" "planton_runner" {
  name       = local.release_name
  repository = local.chart_repository
  chart      = local.helm_chart_name
  version    = local.chart_version
  namespace  = local.namespace

  # The module owns namespace creation (create_namespace flag).
  create_namespace = false

  # No Helm wait — see the module comment (readiness is the work queue,
  # never pod liveness).
  wait    = false
  timeout = 300

  # Documents merged in order by the provider (helm -f semantics): the
  # typed rendering first, the user's escape hatch second — and the
  # enrollment block re-pinned LAST, the one deliberate exception to the
  # escape hatch's last-word contract (twin of the Pulumi module). The
  # token contract (a Secret NAME, never inline token material) is
  # load-bearing; letting an override move it would break the secret
  # discipline.
  values = concat(
    [yamlencode(local.typed_helm_values)],
    try(var.spec.helm_values, "") != "" ? [var.spec.helm_values] : [],
    [yamlencode({ enrollment = local.enrollment_block })]
  )

  lifecycle {
    # FAIL LOUDLY below the enrollment-contract floor: charts before
    # 0.4.0 silently IGNORE the enrollment values — the runner would
    # deploy with no way to join, and nothing downstream would name the
    # cause. Twin: the Pulumi module's Resources() guard.
    precondition {
      condition = (
        tonumber(split(".", local.chart_version)[0]) > tonumber(split(".", local.min_chart_version)[0])
        || (
          tonumber(split(".", local.chart_version)[0]) == tonumber(split(".", local.min_chart_version)[0])
          && (
            tonumber(split(".", local.chart_version)[1]) > tonumber(split(".", local.min_chart_version)[1])
            || (
              tonumber(split(".", local.chart_version)[1]) == tonumber(split(".", local.min_chart_version)[1])
              && tonumber(split(".", local.chart_version)[2]) >= tonumber(split(".", local.min_chart_version)[2])
            )
          )
        )
      )
      error_message = "The chart version predates the runner's token-enrollment contract — use 0.4.0 or newer (charts below it silently ignore the enrollment values)."
    }
  }

  depends_on = [
    kubernetes_namespace_v1.planton_runner,
    kubernetes_secret_v1.runner_token,
  ]
}
