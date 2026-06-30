# 1) Create a certificate for the external/internal hostnames
resource "kubernetes_manifest" "ingress_certificate" {
  count = local.ingress_is_enabled && local.ingress_dns_domain != "" ? 1 : 0

  manifest = yamldecode(
    yamlencode({
      apiVersion = "cert-manager.io/v1"
      kind       = "Certificate"
      metadata = {
        name      = local.certificate_name
        namespace = local.istio_ingress_namespace
        labels    = local.final_labels
      }
      spec = {
        dnsNames   = local.ingress_hostnames
        secretName = local.ingress_cert_secret_name
        issuerRef = {
          kind = "ClusterIssuer"
          name = local.ingress_cert_cluster_issuer_name
        }
      }
    })
  )
}

# 2) Create the external Gateway
resource "kubernetes_manifest" "external_gateway" {
  count = local.ingress_is_enabled && local.ingress_dns_domain != "" ? 1 : 0

  manifest = yamldecode(
    yamlencode({
      apiVersion = "gateway.networking.k8s.io/v1beta1"
      kind       = "Gateway"
      metadata = {
        name      = local.external_gateway_name
        namespace = local.istio_ingress_namespace
        labels    = local.final_labels
      }
      spec = {
        gatewayClassName = local.gateway_ingress_class_name
        addresses = [
          {
            type  = "Hostname"
            value = local.gateway_external_lb_hostname
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
    })
  )

  # To ensure the Certificate exists before the Gateway references it, add an explicit dependency:
  depends_on = [
    kubernetes_manifest.ingress_certificate
  ]
}

# 3) Create an HTTPRoute to redirect external HTTP (port 80) to HTTPS (port 443)
resource "kubernetes_manifest" "http_external_redirect" {
  count = local.ingress_is_enabled && local.ingress_dns_domain != "" ? 1 : 0

  manifest = yamldecode(
    yamlencode({
      apiVersion = "gateway.networking.k8s.io/v1beta1"
      kind       = "HTTPRoute"
      metadata = {
        name      = local.http_external_redirect_route_name
        namespace = local.namespace
        labels    = local.final_labels
      }
      spec = {
        hostnames = [local.ingress_external_hostname]
        parentRefs = [
          {
            name        = local.external_gateway_name
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
    })
  )

  depends_on = [
    kubernetes_manifest.external_gateway
  ]
}

# 4) Create an HTTPRoute to route HTTPS traffic to your Solr service
resource "kubernetes_manifest" "https_external_route" {
  count = local.ingress_is_enabled && local.ingress_dns_domain != "" ? 1 : 0

  manifest = yamldecode(
    yamlencode({
      apiVersion = "gateway.networking.k8s.io/v1beta1"
      kind       = "HTTPRoute"
      metadata = {
        name      = local.https_external_route_name
        namespace = local.namespace
        labels    = local.final_labels
      }
      spec = {
        hostnames = [local.ingress_external_hostname]
        parentRefs = [
          {
            name        = local.external_gateway_name
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
                name      = local.kube_service_name
                namespace = local.namespace
                port      = 80
              }
            ]
          }
        ]
      }
    })
  )

  depends_on = [
    kubernetes_manifest.external_gateway
  ]
}
