locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "strimzi_kafka_operator_kubernetes"
    "resource_name" = var.metadata.name
  }

  org_label = (
    var.metadata.org != null && var.metadata.org != ""
    ? { "organization" = var.metadata.org }
    : {}
  )

  env_label = (
    var.metadata.env != null && var.metadata.env != ""
    ? { "environment" = var.metadata.env }
    : {}
  )

  labels = merge(local.base_labels, local.org_label, local.env_label)

  namespace = coalesce(var.spec.namespace, "strimzi-kafka-operator")

  helm_release_name    = var.metadata.name
  helm_chart_name      = "strimzi-kafka-operator"
  helm_chart_repo      = "https://strimzi.io/charts/"
  helm_chart_version   = "0.42.0"
}
