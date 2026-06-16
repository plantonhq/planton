# Per-database R2 credential Secrets. Mirrors the Pulumi module's r2CredentialEnvVars
# helper (backup_config.go): the access-key id and secret access key are stored in a
# Kubernetes Secret and referenced from the postgresql pod env via secretKeyRef, so the
# credentials never appear in plaintext in the postgresql custom resource or pod spec.

resource "kubernetes_secret_v1" "backup_r2_credentials" {
  count = local.backup_r2 != null ? 1 : 0

  depends_on = [kubernetes_namespace_v1.postgres_namespace]

  metadata {
    name      = local.backup_r2_secret_name
    namespace = local.namespace_name
    labels    = local.final_labels
  }

  type = "Opaque"

  data = {
    access_key_id     = local.backup_r2.access_key_id
    secret_access_key = local.backup_r2.secret_access_key
  }
}

resource "kubernetes_secret_v1" "restore_r2_credentials" {
  count = (local.restore_enabled && local.restore_r2 != null) ? 1 : 0

  depends_on = [kubernetes_namespace_v1.postgres_namespace]

  metadata {
    name      = local.restore_r2_secret_name
    namespace = local.namespace_name
    labels    = local.final_labels
  }

  type = "Opaque"

  data = {
    access_key_id     = local.restore_r2.access_key_id
    secret_access_key = local.restore_r2.secret_access_key
  }
}
