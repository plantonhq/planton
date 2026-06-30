# Local values and computed configuration

locals {
  # Build combined labels
  standard_labels = {
    "managed-by"    = "planton"
    "resource"      = var.metadata.name
    "resource-kind" = "KubernetesSecret"
  }

  labels = merge(local.standard_labels, var.spec.labels)

  # Build annotations
  annotations = var.spec.annotations

  # Determine secret type from which variant is set
  secret_type = (
    var.spec.opaque != null ? "Opaque" :
    var.spec.tls != null ? "kubernetes.io/tls" :
    var.spec.docker_config_json != null ? "kubernetes.io/dockerconfigjson" :
    var.spec.basic_auth != null ? "kubernetes.io/basic-auth" :
    var.spec.ssh_auth != null ? "kubernetes.io/ssh-auth" :
    "Opaque"
  )

  # Build the docker config JSON structure when docker_config_json variant is set
  docker_config_json = var.spec.docker_config_json != null ? jsonencode({
    auths = {
      (var.spec.docker_config_json.registry_server) = {
        username = var.spec.docker_config_json.username
        password = var.spec.docker_config_json.password
        email    = var.spec.docker_config_json.email
        auth     = base64encode("${var.spec.docker_config_json.username}:${var.spec.docker_config_json.password}")
      }
    }
  }) : null

  # Compute secret data map based on which variant is set
  secret_data = (
    var.spec.opaque != null ? var.spec.opaque.data :
    var.spec.tls != null ? {
      "tls.crt" = var.spec.tls.tls_crt
      "tls.key" = var.spec.tls.tls_key
    } :
    var.spec.docker_config_json != null ? {
      ".dockerconfigjson" = local.docker_config_json
    } :
    var.spec.basic_auth != null ? {
      "username" = var.spec.basic_auth.username
      "password" = var.spec.basic_auth.password
    } :
    var.spec.ssh_auth != null ? {
      "ssh-privatekey" = var.spec.ssh_auth.ssh_private_key
    } :
    {}
  )
}
