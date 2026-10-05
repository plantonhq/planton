# The optional HorizontalPodAutoscaler and PodDisruptionBudget (Pulumi twin:
# scaling.go).
resource "kubernetes_horizontal_pod_autoscaler_v2" "flagd" {
  count = try(var.spec.hpa.enabled, false) == true ? 1 : 0
  metadata {
    name      = local.name
    namespace = local.namespace
    labels    = local.labels
  }
  spec {
    min_replicas = try(var.spec.hpa.min_replicas, null) != null ? var.spec.hpa.min_replicas : 1
    max_replicas = try(var.spec.hpa.max_replicas, null) != null ? var.spec.hpa.max_replicas : 10
    scale_target_ref {
      api_version = "apps/v1"
      kind        = "Deployment"
      name        = local.name
    }
    metric {
      type = "Resource"
      resource {
        name = "cpu"
        target {
          type                = "Utilization"
          average_utilization = try(var.spec.hpa.target_cpu_utilization_percent, null) != null ? var.spec.hpa.target_cpu_utilization_percent : 80
        }
      }
    }
    dynamic "metric" {
      for_each = try(var.spec.hpa.target_memory_utilization_percent, null) != null ? [var.spec.hpa.target_memory_utilization_percent] : []
      content {
        type = "Resource"
        resource {
          name = "memory"
          target {
            type                = "Utilization"
            average_utilization = metric.value
          }
        }
      }
    }
  }
  depends_on = [kubernetes_deployment_v1.flagd]
}

resource "kubernetes_pod_disruption_budget_v1" "flagd" {
  count = try(var.spec.pdb.enabled, false) == true ? 1 : 0
  metadata {
    name      = local.name
    namespace = local.namespace
    labels    = local.labels
  }
  spec {
    max_unavailable = try(var.spec.pdb.max_unavailable, "") != "" ? var.spec.pdb.max_unavailable : null
    min_available   = try(var.spec.pdb.max_unavailable, "") != "" ? null : (try(var.spec.pdb.min_available, "") != "" ? var.spec.pdb.min_available : "1")
    selector {
      match_labels = local.selector_labels
    }
  }
  depends_on = [kubernetes_deployment_v1.flagd]
}
