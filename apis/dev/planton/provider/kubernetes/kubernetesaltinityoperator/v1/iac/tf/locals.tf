# Local values for Altinity ClickHouse Operator deployment

locals {
  # Determine the namespace - use from spec or default to "kubernetes-altinity-operator"
  namespace = var.spec.namespace != "" ? var.spec.namespace : "kubernetes-altinity-operator"

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # Format: {metadata.name}-{purpose}
  helm_release_name = var.metadata.name
}

