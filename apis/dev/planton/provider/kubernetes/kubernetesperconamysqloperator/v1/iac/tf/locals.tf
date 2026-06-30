locals {
  # Helm chart configuration - using pxc-operator for production-grade PXC clusters
  helm_chart_name    = "pxc-operator"
  helm_chart_repo    = "https://percona.github.io/percona-helm-charts/"
  helm_chart_version = "1.18.0"

  # Namespace - use from spec or default to "percona-mysql-operator"
  namespace = var.spec.namespace != "" ? var.spec.namespace : "percona-mysql-operator"

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # Format: {metadata.name}-{purpose}
  # Users can prefix metadata.name with component type if needed (e.g., "pxc-operator-prod")
  helm_release_name = "${var.metadata.name}-pxc-operator"

  # Metadata labels
  labels = merge(
    var.metadata.labels != null ? var.metadata.labels : {},
    {
      "app.kubernetes.io/name"       = "percona-mysql-operator"
      "app.kubernetes.io/managed-by" = "planton"
      "app.kubernetes.io/component"  = "database-operator"
    }
  )
}

