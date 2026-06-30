# Conditionally create namespace for Harbor if create_namespace is true
resource "kubernetes_namespace_v1" "harbor_namespace" {
  count = var.harbor_kubernetes.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.labels
  }
}

# Note: Full Terraform implementation would include:
# - Harbor Helm release configuration
# - External database configuration
# - External Redis configuration
# - Storage backend configuration
# - Ingress resources
#
# This is a simplified structure. For production use, implement
# the full Harbor Helm chart deployment similar to the Pulumi module.

