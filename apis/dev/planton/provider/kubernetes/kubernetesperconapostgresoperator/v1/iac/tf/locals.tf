locals {
  # Helm chart configuration - using pg-operator for enterprise-grade PostgreSQL
  helm_chart_name    = "pg-operator"
  helm_chart_repo    = "https://percona.github.io/percona-helm-charts/"
  helm_chart_version = "2.7.0"

  # Namespace - use from spec or default to "kubernetes-percona-postgres-operator"
  namespace = var.spec.namespace != "" ? var.spec.namespace : "kubernetes-percona-postgres-operator"

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # The Helm release name uses metadata.name to ensure uniqueness within the namespace
  helm_release_name = var.metadata.name

  # Metadata labels
  labels = merge(
    var.metadata.labels != null ? var.metadata.labels : {},
    {
      "app.kubernetes.io/name"       = "percona-postgres-operator"
      "app.kubernetes.io/managed-by" = "planton"
      "app.kubernetes.io/component"  = "database-operator"
    }
  )
}

