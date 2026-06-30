locals {
  # Filter secrets to only include those with direct string values (not secret refs)
  string_value_secrets = {
    for s in try(var.spec.container.app.env.secrets, []) :
    s.name => s.value
    if try(s.value, null) != null && s.value != ""
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

  data = { for k, v in local.string_value_secrets : k => base64encode(v) }

  depends_on = [
    kubernetes_namespace.this
  ]
}
