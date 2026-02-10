# Kubernetes Service Terraform Module
# Creates a Kubernetes Service resource with the specified configuration.

resource "kubernetes_service_v1" "service" {
  metadata {
    name        = var.spec.name
    namespace   = var.spec.namespace
    labels      = local.labels
    annotations = local.annotations
  }

  spec {
    type = local.service_type

    # Selector for routing traffic to pods.
    # Not set for ExternalName services or selectorless services.
    selector = length(var.spec.selector) > 0 ? var.spec.selector : null

    # Headless service: set clusterIP to "None".
    cluster_ip = var.spec.headless ? "None" : null

    # ExternalName: the DNS name to proxy to.
    external_name = local.is_external_name ? var.spec.external_dns_name : null

    # External traffic policy: only for NodePort and LoadBalancer.
    external_traffic_policy = local.is_external ? local.external_traffic_policy : null

    # Session affinity.
    session_affinity = local.session_affinity

    # LoadBalancer source ranges for IP-based access control.
    load_balancer_source_ranges = local.service_type == "LoadBalancer" ? var.spec.load_balancer_source_ranges : null

    # Dynamic port blocks from the spec.
    dynamic "port" {
      for_each = var.spec.ports
      content {
        name        = port.value.name != "" ? port.value.name : null
        protocol    = port.value.protocol
        port        = port.value.port
        target_port = port.value.target_port != "" ? port.value.target_port : port.value.port
        node_port   = port.value.node_port > 0 ? port.value.node_port : null
      }
    }
  }
}
