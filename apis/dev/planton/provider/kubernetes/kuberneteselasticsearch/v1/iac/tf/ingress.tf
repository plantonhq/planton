# Only create resources if ingress hostnames are configured
resource "kubernetes_manifest" "ingress_certificate" {
  count = length(local.ingress_hostnames) > 0 ? 1 : 0

  manifest = yamldecode(
    yamlencode({
      apiVersion = "cert-manager.io/v1"
      kind       = "Certificate"
      metadata = {
        name      = local.ingress_certificate_name
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

################################
# Elasticsearch Gateway & Routes
################################

resource "kubernetes_manifest" "elasticsearch_external_gateway" {
  count = local.elasticsearch_ingress_is_enabled && local.elasticsearch_ingress_external_hostname != null ? 1 : 0

  manifest = yamldecode(
    yamlencode({
      apiVersion = "gateway.networking.k8s.io/v1beta1"
      kind       = "Gateway"
      metadata = {
        name      = local.elasticsearch_external_gateway_name
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
            hostname = local.elasticsearch_ingress_external_hostname
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
            hostname = local.elasticsearch_ingress_external_hostname
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

  depends_on = [
    kubernetes_manifest.ingress_certificate
  ]
}

# HTTPRoute to redirect external HTTP->HTTPS
resource "kubernetes_manifest" "elasticsearch_http_external_redirect" {
  count = local.elasticsearch_ingress_is_enabled && local.elasticsearch_ingress_external_hostname != null ? 1 : 0

  manifest = yamldecode(
    yamlencode({
      apiVersion = "gateway.networking.k8s.io/v1beta1"
      kind       = "HTTPRoute"
      metadata = {
        name      = local.elasticsearch_http_redirect_route_name
        namespace = local.namespace_name
        labels    = local.final_labels
      }
      spec = {
        hostnames = [
          local.elasticsearch_ingress_external_hostname
        ]
        parentRefs = [
          {
            name        = local.elasticsearch_external_gateway_name
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
    kubernetes_manifest.elasticsearch_external_gateway
  ]
}

# HTTPRoute to route external HTTPS traffic to Elasticsearch
resource "kubernetes_manifest" "elasticsearch_https_external" {
  count = local.elasticsearch_ingress_is_enabled && local.elasticsearch_ingress_external_hostname != null ? 1 : 0

  manifest = yamldecode(
    yamlencode({
      apiVersion = "gateway.networking.k8s.io/v1beta1"
      kind       = "HTTPRoute"
      metadata = {
        name      = local.elasticsearch_https_route_name
        namespace = local.namespace_name
        labels    = local.final_labels
      }
      spec = {
        hostnames = [
          local.elasticsearch_ingress_external_hostname
        ]
        parentRefs = [
          {
            name        = local.elasticsearch_external_gateway_name
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
                name      = local.elasticsearch_kube_service_name
                namespace = local.namespace_name
                port      = local.elasticsearch_port
              }
            ]
          }
        ]
      }
    })
  )

  depends_on = [
    kubernetes_manifest.elasticsearch_external_gateway
  ]
}

################################
# Kibana Gateway & Routes
################################

resource "kubernetes_manifest" "kibana_external_gateway" {
  count = local.kibana_ingress_is_enabled && local.kibana_ingress_external_hostname != null ? 1 : 0

  manifest = yamldecode(
    yamlencode({
      apiVersion = "gateway.networking.k8s.io/v1beta1"
      kind       = "Gateway"
      metadata = {
        name      = local.kibana_external_gateway_name
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
            hostname = local.kibana_ingress_external_hostname
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
            hostname = local.kibana_ingress_external_hostname
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

  depends_on = [
    kubernetes_manifest.ingress_certificate
  ]
}

# HTTPRoute to redirect external HTTP->HTTPS (Kibana)
resource "kubernetes_manifest" "kibana_http_external_redirect" {
  count = local.kibana_ingress_is_enabled && local.kibana_ingress_external_hostname != null ? 1 : 0

  manifest = yamldecode(
    yamlencode({
      apiVersion = "gateway.networking.k8s.io/v1beta1"
      kind       = "HTTPRoute"
      metadata = {
        name      = local.kibana_http_redirect_route_name
        namespace = local.namespace_name
        labels    = local.final_labels
      }
      spec = {
        hostnames = [
          local.kibana_ingress_external_hostname
        ]
        parentRefs = [
          {
            name        = local.kibana_external_gateway_name
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
    kubernetes_manifest.kibana_external_gateway
  ]
}

# HTTPRoute to route external HTTPS traffic to Kibana
resource "kubernetes_manifest" "kibana_https_external" {
  count = local.kibana_ingress_is_enabled && local.kibana_ingress_external_hostname != null ? 1 : 0

  manifest = yamldecode(
    yamlencode({
      apiVersion = "gateway.networking.k8s.io/v1beta1"
      kind       = "HTTPRoute"
      metadata = {
        name      = local.kibana_https_route_name
        namespace = local.namespace_name
        labels    = local.final_labels
      }
      spec = {
        hostnames = [
          local.kibana_ingress_external_hostname
        ]
        parentRefs = [
          {
            name        = local.kibana_external_gateway_name
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
                name      = local.kibana_kube_service_name
                namespace = local.namespace_name
                port      = local.kibana_port
              }
            ]
          }
        ]
      }
    })
  )

  depends_on = [
    kubernetes_manifest.kibana_external_gateway
  ]
}
