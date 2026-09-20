# Enable the Compute Engine API so a fresh project can host the proxy.
# disable_on_destroy is false: tearing down one proxy must never disable the
# API for everything else in the project.
resource "google_project_service" "compute_api" {
  project = local.project_id
  service = "compute.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# A Compute Engine target HTTP proxy — the plaintext-HTTP frontend adapter
# that binds a forwarding rule (the VIP) to a URL map (the routing brain).
# The proxy is deliberately thin: TLS lives on the target HTTPS proxy
# sibling, routing on the URL map, traffic policy on the backend service.
# The standard production pattern points this proxy at a redirect-only URL
# map (http→https 301) while the HTTPS proxy serves the real application.
#
# GCP models the global and regional proxies as two API collections with
# the same surface (the regional one lacks only proxy_bind, a Traffic
# Director lever). spec.region selects which resource below is created;
# the two blocks mirror each other so a manifest reads the same on either
# scope.
#
# url_map is the only mutable field — GCP repoints it in place via a
# dedicated setUrlMap call, so a live frontend can move to a new routing
# table with no downtime. Everything else (name, description, keep-alive,
# proxy_bind, project, region) is immutable (ForceNew) and briefly breaks
# any forwarding rule referencing the old self_link on recreate.
resource "google_compute_target_http_proxy" "this" {
  count = local.is_regional ? 0 : 1

  name        = local.proxy_name
  project     = local.project_id
  description = local.description

  # Arrives resolved to a literal self-link (or plain name, which the
  # provider expands against the project).
  url_map = var.spec.url_map

  # Only honored by EXTERNAL_MANAGED load balancers; null keeps GCP's
  # default (610s) in charge.
  http_keep_alive_timeout_sec = local.http_keep_alive_timeout_sec

  # Traffic Director binding; null lets the API compute its default (false).
  proxy_bind = local.proxy_bind

  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.compute_api]
}

# The regional twin: the frontend adapter of the regional external and
# regional internal Application Load Balancers. Its URL map must be a
# regional URL map in the same region.
resource "google_compute_region_target_http_proxy" "this" {
  count = local.is_regional ? 1 : 0

  name        = local.proxy_name
  project     = local.project_id
  region      = var.spec.region
  description = local.description

  url_map = var.spec.url_map

  # Immutable on the regional resource (mutable on the global one).
  http_keep_alive_timeout_sec = local.http_keep_alive_timeout_sec

  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.compute_api]
}
