# KubernetesNeo4j Terraform module.
#
# Installs Neo4j from the official Helm chart as a real Helm release. The
# typed spec renders into chart values (locals.helm_values); the admin
# password — declared, or generated here when auth is left empty —
# materializes as the "<name>-auth" Kubernetes Secret the chart consumes via
# neo4j.passwordFromSecret; the helm_values escape hatch
# is passed as a SECOND values document, which the provider merges over the
# first with Helm -f semantics — the exact semantic twin of the Pulumi
# module's buildHelmValues + mergeMaps.
#
# ORDERING IS LOAD-BEARING: the chart looks the passwordFromSecret Secret up
# AT TEMPLATE TIME and fails the install when it is missing, so the auth
# Secret is an explicit dependency of the release — it exists first, always.

# The optional installation namespace. Created before the release; deleted
# with the resource.
resource "kubernetes_namespace_v1" "neo4j" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.labels
  }
}

# The admin password the module mints when auth is left empty. Letters and
# digits, no symbols: the chart splits "neo4j/<password>" on the slash and
# clients embed the value in bolt URLs, and a symbol is the one class that
# can need quoting in either place; 24 alphanumerics with two of each
# class is the catalog's shape for a credential a person may also type.
# Twin: the Pulumi module's RandomPassword.
resource "random_password" "admin" {
  count = local.generate_admin_password ? 1 : 0

  length      = 24
  special     = false
  min_upper   = 2
  min_lower   = 2
  min_numeric = 2

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

# The admin password — declared, or generated above — materialized as an
# Opaque Secret carrying the chart's contract (key NEO4J_AUTH, value
# "neo4j/<password>") AND a bare `password` key for workloads that take a
# password rather than the pair (the password_secret output names it).
# Because neo4j.passwordFromSecret points here, the chart renders no auth
# Secret of its own — this Secret is the only place the credential lands,
# and it never transits chart values. Created for every arm but
# existing_secret, which references a Secret the user owns.
resource "kubernetes_secret_v1" "auth" {
  count = local.create_auth_secret ? 1 : 0

  metadata {
    name      = local.auth_secret_name
    namespace = local.namespace
    labels    = local.labels
  }

  type = "Opaque"

  data = {
    NEO4J_AUTH = sensitive("neo4j/${local.admin_password}")
    password   = sensitive(local.admin_password)
  }

  depends_on = [kubernetes_namespace_v1.neo4j]
}

resource "helm_release" "neo4j" {
  name       = local.release_name
  repository = local.helm_chart_repo
  chart      = local.helm_chart_name
  version    = local.chart_version
  namespace  = local.namespace

  # The module owns namespace creation (create_namespace flag).
  create_namespace = false

  # Wait for the server to become Ready — a database that never starts (bad
  # image, unschedulable pod, unbindable volume, a JVM that OOMs on boot)
  # should fail THIS apply, not the first driver connection. Neo4j
  # recovers/upgrades store files on startup, so the budget is generous.
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
    kubernetes_namespace_v1.neo4j,
    kubernetes_secret_v1.auth,
  ]
}
