# Create a certificate using cert-manager (requires cert-manager CRDs).
resource "kubernetes_manifest" "ingress_certificate" {
  count = local.ingress_is_enabled ? 1 : 0

  manifest = {
    apiVersion = "cert-manager.io/v1"
    kind       = "Certificate"
    metadata = {
      name      = local.ingress_certificate_name
      namespace = "istio-ingress"
      labels    = local.final_labels
    }
    spec = {
      dnsNames = [
        for hostname in [local.ingress_external_hostname, local.ingress_internal_hostname]
        : hostname if hostname != null
      ]
      secretName = local.ingress_cert_secret_name
      issuerRef = {
        kind = "ClusterIssuer"
        name = local.ingress_cert_cluster_issuer_name
      }
    }
  }
}

# Create external gateway (requires Gateway API CRDs).
resource "kubernetes_manifest" "gateway" {
  count = local.ingress_is_enabled ? 1 : 0

  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1beta1"
    kind       = "Gateway"
    metadata = {
      name      = local.external_gateway_name
      namespace = "istio-ingress"
      labels    = local.final_labels
    }
    spec = {
      gatewayClassName = "istio"
      addresses = [
        {
          type  = "Hostname"
          value = "ingress-external.istio-ingress.svc.cluster.local"
        }
      ]
      listeners = [
        {
          name     = "https-external"
          hostname = local.ingress_external_hostname
          port     = 443
          protocol = "HTTPS"
          tls = {
            mode = "Terminate"
            certificateRefs = [
              {
                name = local.ingress_cert_secret_name
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
          hostname = local.ingress_external_hostname
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
    kubernetes_manifest.ingress_certificate
  ]
}

# Create an HTTPRoute that redirects HTTP to HTTPS for the external hostname.
resource "kubernetes_manifest" "http_route_external_redirect" {
  count = local.ingress_is_enabled ? 1 : 0

  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1beta1"
    kind       = "HTTPRoute"
    metadata = {
      name      = local.http_external_redirect_route_name
      namespace = local.namespace
      labels    = local.final_labels
    }
    spec = {
      hostnames = [
        local.ingress_external_hostname
      ]
      parentRefs = [
        {
          name        = local.external_gateway_name
          namespace   = "istio-ingress"
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
    kubernetes_manifest.gateway
  ]
}

# Create an HTTPS route for the external hostname to route traffic to the Locust service.
resource "kubernetes_manifest" "http_route_https_external" {
  count = local.ingress_is_enabled ? 1 : 0

  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1beta1"
    kind       = "HTTPRoute"
    metadata = {
      name      = local.https_external_route_name
      namespace = local.namespace
      labels    = local.final_labels
    }
    spec = {
      hostnames = [
        local.ingress_external_hostname
      ]
      parentRefs = [
        {
          name        = local.external_gateway_name
          namespace   = "istio-ingress"
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
              name      = local.kube_service_name
              namespace = local.namespace
              port      = 80
            }
          ]
        }
      ]
    }
  }

  depends_on = [
    kubernetes_manifest.gateway
  ]
}
