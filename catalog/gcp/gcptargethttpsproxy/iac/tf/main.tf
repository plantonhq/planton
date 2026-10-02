# Enable the Compute Engine API so a fresh project can host the proxy.
# disable_on_destroy is false: tearing down one proxy must never disable the
# API for everything else in the project.
resource "google_project_service" "compute_api" {
  project = local.project_id
  service = "compute.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# A Compute Engine target HTTPS proxy — the TLS-termination node that binds
# a forwarding rule (the VIP) to a URL map (the routing brain) and owns the
# client-facing handshake: certificates, SSL policy, QUIC negotiation, and
# TLS 1.3 early data.
#
# GCP models the global and regional proxies as two API collections that
# share the certificate, URL-map, SSL-policy, and keep-alive surface; the
# regional one lacks the certificate map, QUIC, early-data, and mesh-binding
# levers, which the spec keeps off that arm. spec.region selects which
# resource below is created; the two blocks mirror each other so a manifest
# reads the same on either scope.
#
# Certificates attach through exactly one of three mechanisms (enforced
# pre-deploy by the spec's CEL): the classic ssl_certificates list, the
# certificate_manager_certificates list, or an SNI-scale certificate_map.
# Traffic Director proxies skip certificates and drive TLS through
# server_tls_policy instead.
#
# url_map, the certificate wiring, ssl_policy, server_tls_policy, and
# quic_override update in place via dedicated API calls — certificate
# rotation is attach-new-then-detach-old with zero VIP churn. name,
# description, keep-alive, tls_early_data, proxy_bind, and region are
# immutable (ForceNew) and briefly break any forwarding rule referencing
# the old self_link on recreate.
resource "google_compute_target_https_proxy" "this" {
  count = local.is_regional ? 0 : 1

  name        = local.proxy_name
  project     = local.project_id
  description = local.description

  # Arrives resolved to a literal self-link (or plain name, which the
  # provider expands against the project).
  url_map = var.spec.url_map

  ssl_certificates                 = local.ssl_certificates
  certificate_manager_certificates = local.certificate_manager_certificates
  certificate_map                  = local.certificate_map

  ssl_policy        = local.ssl_policy
  server_tls_policy = local.server_tls_policy

  quic_override  = local.quic_override
  tls_early_data = local.tls_early_data

  # Only honored by EXTERNAL_MANAGED load balancers; null keeps GCP's
  # default (610s) in charge.
  http_keep_alive_timeout_sec = local.http_keep_alive_timeout_sec

  # Traffic Director binding; null lets the API compute its default (false).
  proxy_bind = local.proxy_bind

  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.compute_api]
}

# The regional twin: the TLS front door of the regional external and
# regional internal Application Load Balancers. Its URL map, certificates,
# and SSL policy must all be regional resources in the same region.
resource "google_compute_region_target_https_proxy" "this" {
  count = local.is_regional ? 1 : 0

  name        = local.proxy_name
  project     = local.project_id
  region      = var.spec.region
  description = local.description

  url_map = var.spec.url_map

  ssl_certificates                 = local.ssl_certificates
  certificate_manager_certificates = local.certificate_manager_certificates

  ssl_policy        = local.ssl_policy
  server_tls_policy = local.server_tls_policy

  # Immutable on the regional resource (mutable on the global one).
  http_keep_alive_timeout_sec = local.http_keep_alive_timeout_sec

  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.compute_api]
}
