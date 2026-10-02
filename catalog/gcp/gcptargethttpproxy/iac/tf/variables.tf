variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "GcpTargetHttpProxy specification"
  type = object({
    # The GCP project that owns the target HTTP proxy.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable: changing it destroys and recreates the proxy.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the proxy in GCP. Must be 1-63 characters: lowercase letters,
    # digits, and hyphens; must start with a letter and end with a letter or
    # digit. If not specified, defaults to metadata.name.
    # Immutable: changing it destroys and recreates the proxy, briefly
    # breaking every forwarding rule that references the old self_link.
    proxy_name = optional(string, "")

    # What this proxy fronts and which forwarding rule points at it — write it
    # for the operator tracing a request path later. Immutable.
    description = optional(string, "")

    # The scope selector. Empty builds a GLOBAL target HTTP proxy (the global
    # external ALB, the cross-region internal ALB, Traffic Director); a region
    # name such as us-central1 builds a REGIONAL one (the regional external
    # ALB and the regional internal ALB). The URL map it references must live
    # in the same scope — and, for a regional proxy, the same region — and so
    # must the forwarding rule in front of it. Immutable: a proxy cannot move
    # between scopes or regions.
    region = optional(string, "")

    # The URL map that decides where each request goes — the proxy's single
    # routing dependency. Reference a GcpUrlMap resource or provide a URL map
    # self-link directly. Required. A regional proxy can only point at a
    # regional URL map in its own region (a GcpUrlMap declared with the same
    # region). Mutable: GCP swaps it in place (a dedicated setUrlMap call),
    # so repointing a live frontend at a new routing table causes no
    # downtime.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    url_map = string

    # Seconds an idle client connection is kept open after a response while no
    # matching traffic flows (5-1200). Only honored by load balancers with the
    # EXTERNAL_MANAGED scheme (the envoy-based external ALBs, global and
    # regional), where the GCP default is 610; the classic EXTERNAL ALB
    # ignores it. Raise it above your clients' own keep-alive to avoid the
    # load balancer closing connections first. 0 means unset (GCP applies its
    # default). Immutable on both scopes: changing it destroys and recreates
    # the proxy.
    http_keep_alive_timeout_sec = optional(number, 0)

    # Bind the proxy to the private IPs of the Traffic Director mesh instead
    # of Google's edge. Only meaningful when the forwarding rule that
    # references this proxy uses the INTERNAL_SELF_MANAGED scheme (Traffic
    # Director); leave false for internet-facing load balancers. Global
    # proxies only — Traffic Director has no regional proxy, and the regional
    # resource carries no such argument. Immutable.
    proxy_bind = optional(bool, false)

    # Deletion policy for the proxy — what happens when this resource is
    # destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the proxy is deleted (GCP refuses while a forwarding
    #                rule still references it, so a frontend cannot be torn
    #                down out from under its VIP)
    #   "PREVENT" -- destroy FAILS; protects a production frontend from
    #                accidental teardown
    #   "ABANDON" -- the proxy is removed from management but keeps
    #                serving in GCP
    deletion_policy = optional(string, "")
  })
}
