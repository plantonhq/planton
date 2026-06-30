# Kubernetes Namespace Terraform Module
# This module creates a complete namespace with quotas, policies, and mesh integration

resource "kubernetes_namespace_v1" "namespace" {
  metadata {
    name        = var.spec.name
    labels      = local.labels
    annotations = local.annotations
  }
}

# Resource Quota
resource "kubernetes_resource_quota_v1" "quota" {
  count = local.resource_quota_enabled ? 1 : 0

  metadata {
    name      = "${var.spec.name}-quota"
    namespace = kubernetes_namespace_v1.namespace.metadata[0].name
    labels    = local.labels
  }

  spec {
    hard = local.resource_quota_hard
  }

  depends_on = [kubernetes_namespace_v1.namespace]
}

# Limit Range
resource "kubernetes_limit_range_v1" "limits" {
  count = local.limit_range_enabled ? 1 : 0

  metadata {
    name      = "${var.spec.name}-limits"
    namespace = kubernetes_namespace_v1.namespace.metadata[0].name
    labels    = local.labels
  }

  spec {
    limit {
      type = "Container"
      
      default_request = local.default_requests
      default         = local.default_limits
    }
  }

  depends_on = [kubernetes_namespace_v1.namespace]
}

# Network Policy - Ingress Isolation
resource "kubernetes_network_policy_v1" "ingress" {
  count = local.isolate_ingress ? 1 : 0

  metadata {
    name      = "${var.spec.name}-ingress-policy"
    namespace = kubernetes_namespace_v1.namespace.metadata[0].name
    labels    = local.labels
  }

  spec {
    pod_selector {}
    policy_types = ["Ingress"]

    dynamic "ingress" {
      for_each = local.allowed_ingress_namespaces
      content {
        from {
          namespace_selector {
            match_labels = {
              "kubernetes.io/metadata.name" = ingress.value
            }
          }
        }
      }
    }

    # Allow intra-namespace traffic
    ingress {
      from {
        pod_selector {}
      }
    }
  }

  depends_on = [kubernetes_namespace_v1.namespace]
}

# Network Policy - Egress Restriction
resource "kubernetes_network_policy_v1" "egress" {
  count = local.restrict_egress ? 1 : 0

  metadata {
    name      = "${var.spec.name}-egress-policy"
    namespace = kubernetes_namespace_v1.namespace.metadata[0].name
    labels    = local.labels
  }

  spec {
    pod_selector {}
    policy_types = ["Egress"]

    # Always allow DNS
    egress {
      to {
        namespace_selector {
          match_labels = {
            "kubernetes.io/metadata.name" = "kube-system"
          }
        }
      }
      ports {
        protocol = "UDP"
        port     = "53"
      }
      ports {
        protocol = "TCP"
        port     = "53"
      }
    }

    # Allow egress to specified CIDRs
    dynamic "egress" {
      for_each = local.allowed_egress_cidrs
      content {
        to {
          ip_block {
            cidr = egress.value
          }
        }
      }
    }

    # Allow intra-namespace traffic
    egress {
      to {
        pod_selector {}
      }
    }
  }

  depends_on = [kubernetes_namespace_v1.namespace]
}


