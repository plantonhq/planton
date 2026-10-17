# KubernetesClickHouse Terraform module.
#
# Deploys one operator-managed ClickHouse cluster:
#
#   1. the namespace (optional, create_namespace),
#   2. the auth Secret (when users are declared) — one key per user name,
#      holding the declared password or one generated here for a user
#      declared without one, referenced from the CHI so no password ever
#      appears in the CR,
#   3. the ClickHouseKeeperInstallation CR (when coordination resolves to
#      a managed Keeper) — the same operator reconciles it,
#   4. the clickhouse.altinity.com/v1 ClickHouseInstallation CR itself —
#      host StatefulSets, Services, generated ConfigMaps and the PDB are
#      all operator-created from it. No ingress resources — exposure
#      composes from first-class kinds referencing the exported handles.
#
# The CRs apply through kubectl_manifest (alekc/kubectl): unlike the
# hashicorp provider's kubernetes_manifest resource it needs no cluster
# connection at plan time — a cluster can be PLANNED before the Altinity
# operator's CRDs exist, which is what lets an infra chart deploy the
# operator and its clusters in one run (and lets offline plan proofs work).
#
# No wait_for block, deliberately: cluster readiness depends on the
# operator (image pulls, Keeper quorum, replica sync) that is not part of
# applying the resource — the same never-block-on-a-controller posture as
# the sibling operator-CR modules.
#
# NAMING CONTRACT (operator source at the pinned release): the cluster-wide
# client Service is `clickhouse-<name>`, the per-cluster Service
# `cluster-<name>-<cluster>`, every host's StatefulSet and headless Service
# `chi-<name>-<cluster>-<shard>-<replica>`; the managed Keeper's client
# Service is `keeper-<name>-keeper`.

# The optional namespace. Created before everything else (every other
# resource is namespaced); deleted with the resource. Pre-existing-
# namespace deployments leave create_namespace false.
resource "kubernetes_namespace_v1" "namespace" {
  count = try(var.spec.create_namespace, false) ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.labels
  }
}

# ---- generated passwords ----------------------------------------------------
# One generated password per user declared WITHOUT one, keyed by USER NAME —
# the credential's identity. User names are unique (they are the CHI users
# section's keys), so the password follows its user through spec reorders
# (an index keying would silently SWAP passwords when the list order
# changes), and a renamed user is honestly a NEW credential. A user that
# declares a password is not in this map at all. Twin: the Pulumi module's
# per-user RandomPassword resources.
resource "random_password" "user" {
  for_each = local.generated_password_users

  # Letters and digits, no symbols: clients embed the value in connection
  # URLs (clickhouse://user:password@host), where a symbol is the one class
  # that can need escaping; 32 alphanumerics (~190 bits) is far past any
  # practical bar.
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

# ---- the auth Secret --------------------------------------------------------
# One key per user, value = that user's password: the declared one
# (already resolved) when the spec carries it, the generated one above when
# it does not. The CHI wires each user's password to its key via
# valueFrom.secretKeyRef — rotating the Secret alone does not re-render
# ClickHouse config (upstream-documented); bump any spec field to roll a
# rotation out.
#
# The data is ONE merge: a per-user conditional that indexed
# random_password.user[<name>] would evaluate that lookup for a declared
# user too, whose key is not in the map, and fail the plan.
resource "kubernetes_secret_v1" "auth" {
  count = length(try(var.spec.users, [])) > 0 ? 1 : 0

  metadata {
    name      = local.auth_secret_name
    namespace = local.namespace
    labels    = local.labels
  }

  data = merge(
    { for name, generated in random_password.user : name => generated.result },
    local.declared_passwords,
  )

  type = "Opaque"

  depends_on = [
    kubernetes_namespace_v1.namespace,
  ]
}

# ---- the ClickHouseKeeperInstallation ----------------------------------------
resource "kubectl_manifest" "keeper" {
  count = local.keeper_managed ? 1 : 0

  yaml_body = yamlencode(local.keeper_manifest)

  server_side_apply = true

  depends_on = [
    kubernetes_namespace_v1.namespace,
  ]
}

# ---- the ClickHouseInstallation ----------------------------------------------
# Applied after the Keeper CR so the coordination reference resolves as
# soon as the operator reconciles (apply order, not readiness — see the
# no-wait posture above).
resource "kubectl_manifest" "clickhouse_installation" {
  yaml_body = yamlencode(local.chi_manifest)

  server_side_apply = true

  depends_on = [
    kubernetes_namespace_v1.namespace,
    kubernetes_secret_v1.auth,
    kubectl_manifest.keeper,
  ]
}
