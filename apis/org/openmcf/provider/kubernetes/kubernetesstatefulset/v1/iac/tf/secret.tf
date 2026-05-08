locals {
  # Filter secrets to only include those with direct string values
  string_value_secrets = {
    for k, v in (try(var.spec.container.app.env.secrets, null) != null ? var.spec.container.app.env.secrets : {}) :
    k => v.value
    if try(v.value, null) != null && v.value != ""
  }
}

# Create a secret for environment secrets if any direct string values are defined
resource "kubernetes_secret" "env_secrets" {
  count = length(local.string_value_secrets) > 0 ? 1 : 0

  metadata {
    name      = local.env_secret_name
    namespace = local.namespace
    labels    = local.final_labels
  }

  type = "Opaque"

  data = local.string_value_secrets

  depends_on = [
    kubernetes_namespace.this
  ]
}
