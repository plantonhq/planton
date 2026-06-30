locals {
  # Helm chart configuration
  helm_chart_name    = "psmdb-operator"
  helm_chart_repo    = "https://percona.github.io/percona-helm-charts/"
  helm_chart_version = "1.20.1"

  # Namespace - use from spec or default to "percona-operator"
  namespace = var.spec.namespace != "" ? var.spec.namespace : "percona-operator"

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # Format: {metadata.name} for the Helm release name
  helm_release_name = var.metadata.name

  # Metadata labels
  labels = merge(
    var.metadata.labels != null ? var.metadata.labels : {},
    {
      "app.kubernetes.io/name"       = "percona-mongo-operator"
      "app.kubernetes.io/managed-by" = "planton"
      "app.kubernetes.io/component"  = "database-operator"
    }
  )
}

