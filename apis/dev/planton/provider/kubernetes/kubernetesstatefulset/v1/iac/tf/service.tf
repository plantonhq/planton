# Headless service for stable network identity
# Pod DNS: <pod-name>.<headless-service>.<namespace>.svc.cluster.local
resource "kubernetes_service" "headless" {
  metadata {
    name      = local.headless_service_name
    namespace = local.namespace
    labels    = local.final_labels
  }

  spec {
    type                        = "ClusterIP"
    cluster_ip                  = "None" # Makes it headless
    publish_not_ready_addresses = true   # Important for StatefulSets
    selector                    = local.selector_labels

    dynamic "port" {
      for_each = try(var.spec.container.app.ports, [])
      content {
        name         = port.value.name
        protocol     = port.value.network_protocol
        port         = port.value.service_port
        target_port  = port.value.container_port
        app_protocol = port.value.app_protocol
      }
    }
  }

  depends_on = [
    kubernetes_namespace.this
  ]
}

# ClusterIP service for client access (load-balanced access to pods)
resource "kubernetes_service" "client" {
  count = length(try(var.spec.container.app.ports, [])) > 0 ? 1 : 0

  metadata {
    name      = local.kube_service_name
    namespace = local.namespace
    labels    = local.final_labels
  }

  spec {
    type     = "ClusterIP"
    selector = local.selector_labels

    dynamic "port" {
      for_each = try(var.spec.container.app.ports, [])
      content {
        name         = port.value.name
        protocol     = port.value.network_protocol
        port         = port.value.service_port
        target_port  = port.value.container_port
        app_protocol = port.value.app_protocol
      }
    }
  }

  depends_on = [
    kubernetes_stateful_set.this
  ]
}
