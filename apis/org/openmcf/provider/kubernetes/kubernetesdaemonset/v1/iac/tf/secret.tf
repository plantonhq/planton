##############################################
# Create a Kubernetes Secret for environment secrets that are provided as direct string values.
# Secrets that reference external Kubernetes Secrets (via secret_ref) are not included here;
# they are handled directly in the DaemonSet as environment variable references.
# Uses computed name to avoid conflicts when multiple DaemonSets share a namespace.
##############################################

locals {
  # Filter secrets to only include those with direct string values (not secret refs)
  string_value_secrets = {
    for s in try(var.spec.container.app.env.secrets, []) :
    s.name => s.value
    if try(s.value, null) != null && s.value != ""
  }
}

resource "kubernetes_secret" "this" {
  # Only create the secret if there are direct string values to store
  count = length(local.string_value_secrets) > 0 ? 1 : 0

  metadata {
    name      = local.env_secret_name
    namespace = local.namespace
    labels    = local.final_labels
  }

  type = "Opaque"

  # Populate the secret with key-value pairs (only string values, not secret refs)
  data = { for k, v in local.string_value_secrets : k => base64encode(v) }

  depends_on = [kubernetes_namespace.this]
}

