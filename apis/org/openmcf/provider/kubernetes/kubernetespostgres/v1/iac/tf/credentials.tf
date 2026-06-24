# Per-database R2 credential Secrets: the access-key id and secret access key are stored
# in a Kubernetes Secret and referenced from the postgresql pod env via secretKeyRef, so
# the credentials never appear in plaintext in the postgresql custom resource or pod spec.

resource "kubernetes_secret_v1" "backup_r2_credentials" {
  count = local.backup_creds_present ? 1 : 0

  depends_on = [kubernetes_namespace_v1.postgres_namespace]

  metadata {
    name      = local.backup_r2_secret_name
    namespace = local.namespace_name
    labels    = local.final_labels
  }

  type = "Opaque"

  data = {
    access_key_id     = local.backup_creds.access_key_id
    secret_access_key = local.backup_creds.secret_access_key
  }
}

resource "kubernetes_secret_v1" "restore_r2_credentials" {
  count = local.restore_creds_present ? 1 : 0

  depends_on = [kubernetes_namespace_v1.postgres_namespace]

  metadata {
    name      = local.restore_r2_secret_name
    namespace = local.namespace_name
    labels    = local.final_labels
  }

  type = "Opaque"

  data = {
    access_key_id     = local.restore_creds.access_key_id
    secret_access_key = local.restore_creds.secret_access_key
  }
}
