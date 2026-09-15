# KubernetesPlantonPlatform Terraform module.
#
# Declares one PlantonPlatform custom resource that the Planton operator
# (KubernetesPlantonOperator) reconciles into a running self-hosted
# platform — control plane, console, identity server, databases, secrets
# manager, and in-cluster runner, in one namespace — plus the optional
# owning namespace.
#
# The CR is rendered from null-pruned locals (locals.platform_spec): keys
# render ONLY when the manifest declared them, so the operator's own
# defaulting stays authoritative for everything unset. The exact twin of
# the Pulumi module's apiextensions.NewCustomResource + platformSpecBody.
#
# DESTROY: platform teardown is Kubernetes garbage collection — every
# operator-created object is owner-referenced to the CR, so deletion
# completes even when the operator itself is already gone. The delete
# timeout is headroom, not an expected wait.

# The optional owning namespace. Created before the CR; deleted with the
# resource (pre-existing-namespace installs leave create_namespace false).
resource "kubernetes_namespace_v1" "planton_platform" {
  count = try(var.spec.create_namespace, false) ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.labels
  }
}

# The credentials the database's backup and recovery stores declare,
# materialized as the Secrets the operator's preflight reads
# (`<platform>-postgres-backup-creds` / `-recovery-creds`; keys per backend,
# see locals.object_store_creds_data). A keyless posture creates none and
# the CR names none. Created before the CR so the database is born
# archiving; deleted with the resource. Twin of the Pulumi module's
# createObjectStoreSecrets.
resource "kubernetes_secret_v1" "backup_credentials" {
  count = lookup(local.object_store_creds_data, "backup", null) != null ? 1 : 0

  metadata {
    name      = local.backup_creds_secret_name
    namespace = local.namespace
    labels    = local.labels
  }

  data = local.object_store_creds_data["backup"]

  depends_on = [kubernetes_namespace_v1.planton_platform]
}

resource "kubernetes_secret_v1" "recovery_credentials" {
  count = lookup(local.object_store_creds_data, "recovery", null) != null ? 1 : 0

  metadata {
    name      = local.recovery_creds_secret_name
    namespace = local.namespace
    labels    = local.labels
  }

  data = local.object_store_creds_data["recovery"]

  depends_on = [kubernetes_namespace_v1.planton_platform]
}

# The CA bundle a private S3-compatible endpoint chains to, under the one
# key the CR's endpointCASecretRef names. Only an s3 arm with
# endpoint_ca_pem has one.
resource "kubernetes_secret_v1" "backup_endpoint_ca" {
  count = lookup(local.object_store_ca_data, "backup", null) != null ? 1 : 0

  metadata {
    name      = local.backup_endpoint_ca_name
    namespace = local.namespace
    labels    = local.labels
  }

  data = local.object_store_ca_data["backup"]

  depends_on = [kubernetes_namespace_v1.planton_platform]
}

resource "kubernetes_secret_v1" "recovery_endpoint_ca" {
  count = lookup(local.object_store_ca_data, "recovery", null) != null ? 1 : 0

  metadata {
    name      = local.recovery_endpoint_ca_name
    namespace = local.namespace
    labels    = local.labels
  }

  data = local.object_store_ca_data["recovery"]

  depends_on = [kubernetes_namespace_v1.planton_platform]
}

# The credential a declared vault seal carries (an AWS secret access key, an
# Azure client secret, a transit token), materialized as the Secret the CR
# names (`<platform>-openbao-seal-creds`), keyed by the environment variable
# the seal wrapper reads (see locals.seal_creds_data). A keyless arm — and the
# GCP arm always — creates none and the CR names none. Created before the CR
# so the vault's seal finds its credential at its first start; deleted with
# the resource. Twin of the Pulumi module's seal_secret.go.
resource "kubernetes_secret_v1" "vault_seal_credentials" {
  count = local.seal_creds_data != null ? 1 : 0

  metadata {
    name      = local.seal_creds_secret_name
    namespace = local.namespace
    labels    = local.labels
  }

  data = local.seal_creds_data

  depends_on = [kubernetes_namespace_v1.planton_platform]
}

# The PlantonPlatform declaration.
resource "kubectl_manifest" "planton_platform" {
  yaml_body = yamlencode({
    apiVersion = local.api_version
    kind       = local.cr_kind
    metadata = {
      # Namespaced, named from THIS resource's metadata.name — the prefix
      # of every object the operator creates for the platform.
      name      = local.platform_name
      namespace = local.namespace
      labels    = local.labels
    }
    spec = local.platform_spec
  })

  server_side_apply = true
  force_conflicts   = true

  timeouts {
    delete = "15m"
  }

  depends_on = [
    kubernetes_namespace_v1.planton_platform,
    kubernetes_secret_v1.backup_credentials,
    kubernetes_secret_v1.recovery_credentials,
    kubernetes_secret_v1.backup_endpoint_ca,
    kubernetes_secret_v1.recovery_endpoint_ca,
    kubernetes_secret_v1.vault_seal_credentials,
  ]
}
