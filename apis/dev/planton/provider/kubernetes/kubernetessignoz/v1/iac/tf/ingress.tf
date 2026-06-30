##############################################
# ingress.tf
#
# Creates cert-manager Certificate and Gateway
# API resources for SigNoz UI and OTel Collector
# if ingress is enabled.
##############################################

#####################################
# SigNoz UI Ingress Resources
#####################################

# Create a certificate for SigNoz UI using cert-manager
resource "kubernetes_manifest" "signoz_certificate" {
  count = local.signoz_ingress_is_enabled && local.signoz_ingress_external_hostname != null ? 1 : 0

  manifest = {
    apiVersion = "cert-manager.io/v1"
    kind       = "Certificate"
    metadata = {
      name      = local.signoz_certificate_name
      namespace = local.istio_ingress_namespace
      labels    = local.final_labels
    }
    spec = {
      dnsNames   = [local.signoz_ingress_external_hostname]
      secretName = local.signoz_certificate_name
      issuerRef = {
        kind = "ClusterIssuer"
        name = local.signoz_cert_cluster_issuer_name
      }
    }
  }

  depends_on = [
    kubernetes_namespace_v1.signoz_namespace
  ]
}

# Create Gateway for SigNoz UI external access
resource "kubernetes_manifest" "signoz_gateway" {
  count = local.signoz_ingress_is_enabled && local.signoz_ingress_external_hostname != null ? 1 : 0

  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "Gateway"
    metadata = {
      name      = local.signoz_gateway_name
      namespace = local.istio_ingress_namespace
      labels    = local.final_labels
    }
    spec = {
      gatewayClassName = "istio"
      addresses = [
        {
          type  = "Hostname"
          value = local.gateway_external_loadbalancer_service_hostname
        }
      ]
      listeners = [
        # HTTPS listener for secure SigNoz UI access
        {
          name     = "https-external"
          hostname = local.signoz_ingress_external_hostname
          port     = 443
          protocol = "HTTPS"
          tls = {
            mode = "Terminate"
            certificateRefs = [
              {
                name = local.signoz_certificate_name
              }
            ]
          }
          allowedRoutes = {
            namespaces = {
              from = "All"
            }
          }
        },
        # HTTP listener for HTTP-to-HTTPS redirect
        {
          name     = "http-external"
          hostname = local.signoz_ingress_external_hostname
          port     = 80
          protocol = "HTTP"
          allowedRoutes = {
            namespaces = {
              from = "All"
            }
          }
        }
      ]
    }
  }

  depends_on = [
    kubernetes_manifest.signoz_certificate
  ]
}

# HTTPRoute for HTTP-to-HTTPS redirect (SigNoz UI)
resource "kubernetes_manifest" "signoz_http_redirect" {
  count = local.signoz_ingress_is_enabled && local.signoz_ingress_external_hostname != null ? 1 : 0

  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "HTTPRoute"
    metadata = {
      name      = local.signoz_http_redirect_route_name
      namespace = local.namespace
      labels    = local.final_labels
    }
    spec = {
      hostnames = [local.signoz_ingress_external_hostname]
      parentRefs = [
        {
          name        = local.signoz_gateway_name
          namespace   = local.istio_ingress_namespace
          sectionName = "http-external"
        }
      ]
      rules = [
        {
          filters = [
            {
              type = "RequestRedirect"
              requestRedirect = {
                scheme     = "https"
                statusCode = 301
              }
            }
          ]
        }
      ]
    }
  }

  depends_on = [
    kubernetes_manifest.signoz_gateway
  ]
}

# HTTPRoute for HTTPS traffic to SigNoz frontend service
resource "kubernetes_manifest" "signoz_https_route" {
  count = local.signoz_ingress_is_enabled && local.signoz_ingress_external_hostname != null ? 1 : 0

  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "HTTPRoute"
    metadata = {
      name      = local.signoz_https_route_name
      namespace = local.namespace
      labels    = local.final_labels
    }
    spec = {
      hostnames = [local.signoz_ingress_external_hostname]
      parentRefs = [
        {
          name        = local.signoz_gateway_name
          namespace   = local.istio_ingress_namespace
          sectionName = "https-external"
        }
      ]
      rules = [
        {
          matches = [
            {
              path = {
                type  = "PathPrefix"
                value = "/"
              }
            }
          ]
          backendRefs = [
            {
              # Route to the frontend service created by SigNoz Helm chart
              name      = "${var.metadata.name}-frontend"
              namespace = local.namespace
              port      = local.signoz_frontend_port
            }
          ]
        }
      ]
    }
  }

  depends_on = [
    kubernetes_manifest.signoz_gateway
  ]
}

#####################################
# OTel Collector Ingress Resources
#####################################

# Create a certificate for OTel Collector using cert-manager
resource "kubernetes_manifest" "otel_certificate" {
  count = local.otel_collector_ingress_is_enabled && local.otel_collector_external_http_hostname != null ? 1 : 0

  manifest = {
    apiVersion = "cert-manager.io/v1"
    kind       = "Certificate"
    metadata = {
      name      = local.otel_certificate_name
      namespace = local.istio_ingress_namespace
      labels    = local.final_labels
    }
    spec = {
      dnsNames   = [local.otel_collector_external_http_hostname]
      secretName = local.otel_certificate_name
      issuerRef = {
        kind = "ClusterIssuer"
        name = local.otel_cert_cluster_issuer_name
      }
    }
  }

  depends_on = [
    kubernetes_namespace_v1.signoz_namespace
  ]
}

# Create Gateway for OTel Collector external access
resource "kubernetes_manifest" "otel_gateway" {
  count = local.otel_collector_ingress_is_enabled && local.otel_collector_external_http_hostname != null ? 1 : 0

  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "Gateway"
    metadata = {
      name      = local.otel_gateway_name
      namespace = local.istio_ingress_namespace
      labels    = local.final_labels
    }
    spec = {
      gatewayClassName = "istio"
      addresses = [
        {
          type  = "Hostname"
          value = local.gateway_external_loadbalancer_service_hostname
        }
      ]
      listeners = [
        # HTTPS listener for OTel Collector HTTP endpoint
        {
          name     = "https-otel-http"
          hostname = local.otel_collector_external_http_hostname
          port     = 443
          protocol = "HTTPS"
          tls = {
            mode = "Terminate"
            certificateRefs = [
              {
                name = local.otel_certificate_name
              }
            ]
          }
          allowedRoutes = {
            namespaces = {
              from = "All"
            }
          }
        },
        # HTTP listener for HTTP-to-HTTPS redirect
        {
          name     = "http-otel-http"
          hostname = local.otel_collector_external_http_hostname
          port     = 80
          protocol = "HTTP"
          allowedRoutes = {
            namespaces = {
              from = "All"
            }
          }
        }
      ]
    }
  }

  depends_on = [
    kubernetes_manifest.otel_certificate
  ]
}

# HTTPRoute for HTTP-to-HTTPS redirect (OTel Collector)
resource "kubernetes_manifest" "otel_http_redirect" {
  count = local.otel_collector_ingress_is_enabled && local.otel_collector_external_http_hostname != null ? 1 : 0

  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "HTTPRoute"
    metadata = {
      name      = local.otel_http_redirect_route_name
      namespace = local.namespace
      labels    = local.final_labels
    }
    spec = {
      hostnames = [local.otel_collector_external_http_hostname]
      parentRefs = [
        {
          name        = local.otel_gateway_name
          namespace   = local.istio_ingress_namespace
          sectionName = "http-otel-http"
        }
      ]
      rules = [
        {
          filters = [
            {
              type = "RequestRedirect"
              requestRedirect = {
                scheme     = "https"
                statusCode = 301
              }
            }
          ]
        }
      ]
    }
  }

  depends_on = [
    kubernetes_manifest.otel_gateway
  ]
}

# HTTPRoute for HTTPS traffic to OTel Collector HTTP endpoint
resource "kubernetes_manifest" "otel_https_route" {
  count = local.otel_collector_ingress_is_enabled && local.otel_collector_external_http_hostname != null ? 1 : 0

  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "HTTPRoute"
    metadata = {
      name      = local.otel_https_route_name
      namespace = local.namespace
      labels    = local.final_labels
    }
    spec = {
      hostnames = [local.otel_collector_external_http_hostname]
      parentRefs = [
        {
          name        = local.otel_gateway_name
          namespace   = local.istio_ingress_namespace
          sectionName = "https-otel-http"
        }
      ]
      rules = [
        {
          matches = [
            {
              path = {
                type  = "PathPrefix"
                value = "/"
              }
            }
          ]
          backendRefs = [
            {
              # Route to OTel Collector HTTP port (4318)
              name      = "${var.metadata.name}-otel-collector"
              namespace = local.namespace
              port      = local.otel_http_port
            }
          ]
        }
      ]
    }
  }

  depends_on = [
    kubernetes_manifest.otel_gateway
  ]
}

