# KubernetesValkey Terraform module.
#
# Installs Valkey from the official Helm chart as a real Helm release. The
# typed spec renders into chart values (locals.helm_values); ACL passwords —
# declared, or generated here for every user declared without one —
# materialize as the "<name>-auth" Kubernetes Secret the chart consumes via
# auth.usersExistingSecret; the helm_values escape hatch is
# passed as a SECOND values document, which the provider merges over the
# first with Helm -f semantics — the exact semantic twin of the Pulumi
# module's buildHelmValues + mergeMaps.
#
# The release is named after metadata.name (NOT a fixed chart name) and the
# chart's fullname is pinned to the same value: several Valkey instances
# coexist in one cluster, each rendering its own `<name>`,
# `<name>-headless`, and (replication) `<name>-read` Services.

# The optional installation namespace. Created before the release; deleted
# with the resource.
resource "kubernetes_namespace_v1" "valkey" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.labels
  }
}

# One generated password per ACL user declared WITHOUT one, keyed by
# USERNAME — the credential's identity. Usernames are unique across the
# auth block (CEL-enforced), so the password follows its user through spec
# reorders (an index keying would silently SWAP passwords when the list
# order changes), and a renamed user is honestly a NEW credential. A user
# that declares a password is not in this map at all. Twin: the Pulumi
# module's per-username RandomPassword resources.
resource "random_password" "user" {
  for_each = local.generated_password_users

  # Letters and digits, no symbols: the chart's init script reads the value
  # from the mounted Secret and clients embed it in connection URLs, and a
  # symbol is the one class that can need quoting in either place; 32
  # alphanumerics (~190 bits) is far past any practical bar.
  length  = 32
  special = false

  # The generation-shape arguments are ignored after creation so an
  # IMPORTED credential never silently regenerates: rotation stays an
  # explicit verb, never plan fallout. Twin: the Pulumi module's
  # IgnoreChanges on the same argument set.
  lifecycle {
    ignore_changes = [
      length, special, upper, lower, numeric,
      min_lower, min_numeric, min_special, min_upper, override_special,
    ]
  }
}

# The ACL passwords, materialized as an Opaque Secret — ONE KEY PER
# USERNAME, each key's value that user's password (declared, or generated
# above). That layout is the chart's contract for auth.usersExistingSecret:
# its init script reads /valkey-users-secret/<passwordKey> where passwordKey
# defaults to the username (the module leaves passwordKey unset), and its
# metrics exporter reads the "default" key the same way. Because the
# rendered aclUsers carry no inline passwords, the chart renders no auth
# Secret of its own — this Secret is the only place the credentials land,
# and it never transits chart values.
#
# The data is ONE merge, generated under declared: a per-user conditional
# that indexed random_password.user[<name>] would evaluate that lookup for
# a declared user too, whose key is not in the map, and fail the plan.
resource "kubernetes_secret_v1" "auth" {
  count = local.auth_enabled ? 1 : 0

  metadata {
    name      = local.auth_secret_name
    namespace = local.namespace
    labels    = local.labels
  }

  type = "Opaque"

  data = merge(
    { for name, generated in random_password.user : name => generated.result },
    local.declared_passwords,
  )

  depends_on = [kubernetes_namespace_v1.valkey]
}

resource "helm_release" "valkey" {
  name       = local.release_name
  repository = local.helm_chart_repo
  chart      = local.helm_chart_name
  version    = local.chart_version
  namespace  = local.namespace

  # The module owns namespace creation (create_namespace flag).
  create_namespace = false

  # Wait for the workload to become Ready — a store that never starts (bad
  # image, unschedulable pod, unbindable volume) should fail THIS apply,
  # not the first client connection. Replication starts pods one at a time
  # (OrderedReady) and each replica full-syncs before Ready, so the budget
  # is sized for a multi-pod StatefulSet, not a single Deployment.
  wait            = true
  atomic          = true
  cleanup_on_fail = true
  timeout         = 600

  # Two documents, merged in order by the provider (helm -f semantics):
  # the typed rendering first, the user's escape hatch last.
  values = [
    yamlencode(local.helm_values),
    try(var.spec.helm_values, ""),
  ]

  depends_on = [
    kubernetes_namespace_v1.valkey,
    kubernetes_secret_v1.auth,
  ]
}
