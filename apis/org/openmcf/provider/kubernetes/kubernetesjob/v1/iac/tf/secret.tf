##############################################
# Create a Kubernetes Secret for KubernetesJob
##############################################

locals {
  # Filter secrets to only include those with direct string values
  string_value_secrets = {
    for s in try(var.spec.env.secrets, []) :
    s.name => s.value
    if try(s.value, null) != null && s.value != ""
  }
}

resource "kubernetes_secret" "this" {
  # Only create if there are direct string values
  count = length(local.string_value_secrets) > 0 ? 1 : 0

  metadata {
    # Computed name to avoid conflicts when multiple instances share a namespace
    name      = local.env_secrets_secret_name
    namespace = local.namespace_name
    labels    = local.final_labels
  }

  type = "Opaque"

  # `data` automatically converts each map value into a string,
  # then Kubernetes encodes it as base64 in the final secret.
  data = local.string_value_secrets
}
