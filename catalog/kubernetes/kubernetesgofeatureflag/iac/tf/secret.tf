# Every secret value of the spec (API keys, tokens, passwords, webhook URLs,
# credential headers) as the module-owned `<metadata.name>-env` Opaque
# Secret, one data key per environment variable name (Pulumi twin:
# env_secret.go). The chart's env values reference it by secretKeyRef, so the
# relay configuration ConfigMap and the Helm values carry no secret. Created
# only when the spec holds a secret.
resource "kubernetes_secret_v1" "env" {
  count = local.env_secret_name != "" ? 1 : 0

  metadata {
    name      = local.env_secret_name
    namespace = local.namespace
    labels    = local.labels
  }

  type = "Opaque"
  data = local.secret_env

  depends_on = [kubernetes_namespace_v1.relay]
}
