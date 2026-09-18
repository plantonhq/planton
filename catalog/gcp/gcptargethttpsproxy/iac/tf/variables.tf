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
  description = "GcpTargetHttpsProxy specification"
  type = object({
    # The GCP project that owns the target HTTPS proxy.
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
    # for the operator tracing a TLS incident later. Immutable.
    description = optional(string, "")

    # The URL map that decides where each decrypted request goes — the proxy's
    # single routing dependency. Reference a GcpUrlMap resource or provide a
    # URL map self-link directly. Required. Mutable: GCP swaps it in place (a
    # dedicated setUrlMap call), so repointing a live frontend at a new
    # routing table causes no downtime.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    url_map = string

    # Compute Engine SSL certificates presented to clients (1-15). Reference
    # GcpManagedSslCertificate resources (the default kind), self-managed
    # GcpSslCertificate resources via an explicit valueFrom.kind, or provide
    # SSL certificate self-links directly — both certificate kinds share one
    # API collection and attach identically. The load balancer picks the
    # certificate matching the client's SNI hostname.
    # Not honored by Traffic Director (INTERNAL_SELF_MANAGED) proxies — use
    # server_tls_policy there. Mutually exclusive with
    # certificate_manager_certificates and certificate_map. Mutable: GCP swaps
    # the list in place (setSslCertificates), which is how zero-downtime
    # certificate rotation works — attach the replacement before detaching the
    # old one.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    ssl_certificates = optional(list(string), [])

    # Certificate Manager certificates presented to clients — only honored by
    # the cross-region internal ALB (INTERNAL_MANAGED); external ALBs use
    # certificate_map instead. Reference GcpCertManagerCert resources or
    # provide certificate resource names directly, in the form
    # projects/{project}/locations/{location}/certificates/{name} (a
    # //certificatemanager.googleapis.com/ prefix is also accepted). Mutually
    # exclusive with ssl_certificates and certificate_map. Mutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    certificate_manager_certificates = optional(list(string), [])

    # A Certificate Manager certificate map that selects the served
    # certificate by SNI hostname — the mechanism for serving many domains
    # (SaaS custom domains) beyond the 15-certificate list limit. Only honored
    # by external ALBs (EXTERNAL / EXTERNAL_MANAGED). Format:
    # //certificatemanager.googleapis.com/projects/{project}/locations/{location}/certificateMaps/{name}.
    # Mutually exclusive with ssl_certificates and
    # certificate_manager_certificates. Mutable.
    certificate_map = optional(string, "")

    # The SSL policy constraining TLS versions and cipher suites for client
    # handshakes. Reference a GcpSslPolicy resource or provide an SSL policy
    # self-link directly (e.g.
    # https://www.googleapis.com/compute/v1/projects/{project}/global/sslPolicies/{name}).
    # If not set, GCP applies its permissive default policy (min TLS 1.0,
    # COMPATIBLE profile) — set one to enforce modern TLS for compliance.
    # Mutable: GCP swaps it in place (setSslPolicy).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    ssl_policy = optional(string, "")

    # A network security ServerTlsPolicy resource that configures server-side
    # TLS — the mTLS mechanism: it can demand and validate client
    # certificates. Applies to global proxies behind EXTERNAL /
    # EXTERNAL_MANAGED / INTERNAL_SELF_MANAGED forwarding rules; for Traffic
    # Director this is the ONLY TLS lever (ssl_certificates are ignored).
    # Format: projects/{project}/locations/global/serverTlsPolicies/{name}.
    # If left blank, no server-side TLS policy applies. Mutable — and
    # clearable: removing it PATCHes the proxy back to no policy.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    server_tls_policy = optional(string, "")

    # QUIC (HTTP/3) negotiation policy. NONE lets Google decide (currently
    # enables QUIC), ENABLE forces QUIC negotiation on, DISABLE turns it off.
    # GCP default: NONE. Mutable.
    quic_override = optional(string)

    # TLS 1.3 0-RTT "early data" policy — lets a resuming client send the
    # first HTTP request inside the TLS handshake itself (zero effective round
    # trips, over TCP and QUIC/HTTP-3). Early data is replayable by design, so
    # the modes trade latency against replay safety: STRICT accepts it only
    # for safe methods (GET/HEAD) with no query parameters, PERMISSIVE for all
    # requests, UNRESTRICTED additionally skips rejecting non-idempotent
    # replays (only for services that tolerate replays), DISABLED turns it
    # off. Empty lets GCP apply its default (DISABLED). Immutable: changing it
    # destroys and recreates the proxy.
    tls_early_data = optional(string, "")

    # Seconds an idle client connection is kept open after a response while no
    # matching traffic flows (5-1200). Only honored by load balancers with the
    # EXTERNAL_MANAGED scheme (the envoy-based global external ALB), where the
    # GCP default is 610; the classic EXTERNAL ALB ignores it. Raise it above
    # your clients' own keep-alive to avoid the load balancer closing
    # connections first. 0 means unset (GCP applies its default). Immutable:
    # changing it destroys and recreates the proxy.
    http_keep_alive_timeout_sec = optional(number, 0)

    # Bind the proxy to the private IPs of the Traffic Director mesh instead
    # of Google's edge. Only meaningful when the forwarding rule that
    # references this proxy uses the INTERNAL_SELF_MANAGED scheme (Traffic
    # Director); leave false for internet-facing load balancers. Immutable.
    proxy_bind = optional(bool, false)

    # Deletion policy for the proxy — what happens when this resource is
    # destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the proxy is deleted (GCP refuses while a forwarding
    #                rule still references it, so a TLS frontend cannot be
    #                torn down out from under its VIP)
    #   "PREVENT" -- destroy FAILS; protects a production TLS frontend from
    #                accidental teardown
    #   "ABANDON" -- the proxy is removed from management but keeps
    #                terminating TLS in GCP
    deletion_policy = optional(string, "")
  })
}
