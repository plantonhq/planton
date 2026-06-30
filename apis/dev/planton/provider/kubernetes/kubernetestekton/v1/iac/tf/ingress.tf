##############################################
# ingress.tf
#
# Gateway API resources for Tekton Dashboard ingress.
# Creates: Certificate → Gateway → HTTPRoutes
#
# Pattern follows existing Planton components.
##############################################

##############################################
# Certificate (cert-manager)
##############################################
resource "kubernetes_manifest" "dashboard_certificate" {
  count = local.ingress_enabled && local.dashboard_enabled ? 1 : 0

  manifest = {
    apiVersion = "cert-manager.io/v1"
    kind       = "Certificate"
    metadata = {
      name      = local.cert_secret_name
      namespace = local.istio_ingress_namespace
      labels    = local.final_labels
    }
    spec = {
      dnsNames   = [local.ingress_hostname]
      secretName = local.cert_secret_name
      issuerRef = {
        kind = "ClusterIssuer"
        name = local.cluster_issuer_name
      }
    }
  }

  depends_on = [kubectl_manifest.tekton_dashboard]
}

##############################################
# Gateway (Kubernetes Gateway API)
##############################################
resource "kubernetes_manifest" "dashboard_gateway" {
  count = local.ingress_enabled && local.dashboard_enabled ? 1 : 0

  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "Gateway"
    metadata = {
      name      = local.gateway_name
      namespace = local.istio_ingress_namespace
      labels    = local.final_labels
    }
    spec = {
      gatewayClassName = local.gateway_ingress_class_name
      addresses = [
        {
          type  = "Hostname"
          value = local.gateway_external_lb_service_hostname
        }
      ]
      listeners = [
        {
          name     = "https-external"
          hostname = local.ingress_hostname
          port     = 443
          protocol = "HTTPS"
          tls = {
            mode = "Terminate"
            certificateRefs = [
              {
                name = local.cert_secret_name
              }
            ]
          }
          allowedRoutes = {
            namespaces = {
              from = "All"
            }
          }
        },
        {
          name     = "http-external"
          hostname = local.ingress_hostname
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

  depends_on = [kubernetes_manifest.dashboard_certificate]
}

##############################################
# HTTPRoute (HTTP → HTTPS redirect)
##############################################
resource "kubernetes_manifest" "dashboard_http_redirect_route" {
  count = local.ingress_enabled && local.dashboard_enabled ? 1 : 0

  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "HTTPRoute"
    metadata = {
      name      = local.http_redirect_route_name
      namespace = local.namespace
      labels    = local.final_labels
    }
    spec = {
      hostnames = [local.ingress_hostname]
      parentRefs = [
        {
          name        = local.gateway_name
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

  depends_on = [kubernetes_manifest.dashboard_gateway]
}

##############################################
# HTTPRoute (HTTPS → Dashboard backend)
##############################################
resource "kubernetes_manifest" "dashboard_https_route" {
  count = local.ingress_enabled && local.dashboard_enabled ? 1 : 0

  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "HTTPRoute"
    metadata = {
      name      = local.https_route_name
      namespace = local.namespace
      labels    = local.final_labels
    }
    spec = {
      hostnames = [local.ingress_hostname]
      parentRefs = [
        {
          name        = local.gateway_name
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
              name      = local.dashboard_service_name
              namespace = local.namespace
              port      = local.dashboard_service_port
            }
          ]
        }
      ]
    }
  }

  depends_on = [kubernetes_manifest.dashboard_gateway]
}
