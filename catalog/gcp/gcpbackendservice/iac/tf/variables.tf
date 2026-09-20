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
  description = "GcpBackendService specification"
  type = object({
    # The GCP project that owns the backend service.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable: changing it destroys and recreates the backend service.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the backend service in GCP. Must be 1-63 characters: lowercase
    # letters, digits, and hyphens; must start with a letter and end with a
    # letter or digit. If not specified, defaults to metadata.name.
    # Immutable: changing it destroys and recreates the backend service,
    # briefly breaking every URL map that references the old self_link.
    backend_service_name = optional(string, "")

    # What this backend service fronts and which URL maps route to it — write
    # it for the operator tracing a request path later. Mutable.
    description = optional(string, "")

    # The scope selector. Empty builds a GLOBAL backend service (the global
    # external ALB, the cross-region internal ALB, Traffic Director); a region
    # name such as us-central1 builds a REGIONAL one (the regional external
    # and internal ALBs, and the internal and external passthrough Network
    # Load Balancers). A regional backend service takes a regional health
    # check for the ALB schemes, is routed to only by regional URL maps and
    # regional forwarding rules, and attaches only a regional Cloud Armor
    # policy. The passthrough levers (network, backend failover,
    # failover_policy, connection_tracking_policy, ha_policy,
    # network_pass_through_lb_traffic_policy, the INTERNAL scheme) exist only
    # here and are rejected when region is empty; the global edge levers
    # (compression_mode, custom request/response headers, edge_security_policy,
    # service_lb_policy, the EXTERNAL_MANAGED migration canary, backend
    # preference, locality_lb_policies, max_stream_duration,
    # security_settings, signed_url_keys, and the CDN knobs the regional
    # resource lacks) are rejected when it is set. Immutable: a backend
    # service cannot move between scopes or regions.
    region = optional(string, "")

    # The protocol the load balancer uses to talk to the backends (default
    # HTTP). This is the LB→backend leg, independent of what clients speak to
    # the load balancer: an HTTPS frontend commonly forwards to HTTP backends.
    # H2C is HTTP/2 over cleartext. Must be GRPC when the backend service is
    # referenced by a URL map bound to a target gRPC proxy. For the
    # passthrough Network Load Balancers (regional, scheme INTERNAL or
    # EXTERNAL) use TCP, UDP, or UNSPECIFIED — UNSPECIFIED forwards every IP
    # protocol and is what a forwarding rule with ip_protocol L3_DEFAULT
    # requires. Mutable, but switching protocol families usually also means
    # changing the health check and backend ports.
    protocol = optional(string)

    # Which load balancer family this backend service serves (default
    # EXTERNAL on both scopes: the classic global external Application LB, or
    # the backend-service-based external passthrough Network Load Balancer on
    # a regional service). EXTERNAL_MANAGED is the envoy-based external ALB
    # (global, or regional with region set); INTERNAL_MANAGED is the internal
    # ALB (cross-region on a global service, regional with region set);
    # INTERNAL — regional services only — is the internal passthrough Network
    # Load Balancer; INTERNAL_SELF_MANAGED is Traffic Director / service mesh.
    # Both engines send EXTERNAL explicitly when this is left empty, on both
    # scopes, so an unset scheme means the same thing wherever the service
    # lives (Google's own default differs per scope: EXTERNAL_MANAGED
    # globally, INTERNAL regionally). Regional presets therefore name the
    # scheme outright. A backend service created for one family cannot serve
    # another — the only in-place transition GCP supports is the canary
    # migration EXTERNAL → EXTERNAL_MANAGED driven by
    # external_managed_migration_state, on the global service.
    load_balancing_scheme = optional(string)

    # Name of the backend port to use for instance-group backends. The same
    # named port must be defined on every instance group this service
    # references — each group maps the logical name to its own port number.
    # Required by GCP when the scheme is EXTERNAL and the backends are
    # instance groups; ignored for NEG backends (endpoints carry their own
    # ports). Mutable.
    port_name = optional(string, "")

    # Seconds the load balancer waits for a backend to fully respond before
    # giving up on the request (default 30). For streaming workloads
    # (WebSockets, gRPC streams, long polling) raise this well above the
    # longest expected stream duration. Not used by serverless NEG backends —
    # Cloud Run/Functions manage their own request timeouts. Mutable.
    timeout_sec = optional(number)

    # Seconds an instance being removed or unhealthy keeps its existing
    # connections open to finish in-flight requests (default 300). Lower it
    # for fast-draining stateless services; raise it for long-lived
    # connections. Mutable.
    connection_draining_timeout_sec = optional(number)

    # The health check that decides which backends receive traffic. GCP
    # allows at most ONE health check per backend service, so this is a
    # single reference, not a list. Reference a GcpHealthCheck resource or
    # provide a health check self-link directly. Required by GCP unless every
    # backend is an internet or serverless NEG — serverless platforms manage
    # their own health. A regional backend service behind an Application Load
    # Balancer needs a REGIONAL health check in its own region (a
    # GcpHealthCheck declared with the same region); the passthrough Network
    # Load Balancers accept a global or regional one. Not allowed together
    # with ha_policy. Mutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    health_check = optional(string, "")

    # The VPC network the backends live in — used by the internal passthrough
    # Network Load Balancer, and by an external passthrough one only when it
    # carries an ha_policy with fast IP move. Reference a GcpVpcNetwork
    # resource or provide a network self-link. Regional backend services
    # only. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    network = optional(string, "")

    # The backends that actually serve traffic — instance groups or network
    # endpoint groups, each with its own balancing mode and capacity dials.
    # A backend service may mix backends of the same family but cannot mix
    # instance groups with NEGs. May be empty: a backend service with only a
    # health check is valid and is the natural creation order before instance
    # groups or NEGs exist. Mutable — adding and removing backends is the
    # normal scaling/blue-green operation.
    backends = optional(list(object({
      # Fully-qualified URL of the instance group or network endpoint group
      # serving this backend. Accepts an instance group (zonal or regional) or
      # a NEG self-link; all backends of one service must be the same family —
      # GCP rejects mixing instance groups with NEGs. Provide the URL directly
      # or reference the resource that owns it: the default reference kind is a
      # GcpRegionNetworkEndpointGroup (the serverless/PSC/internet backend
      # bridge), but any group producer can be referenced explicitly by kind.
      # For NEG backends GCP ignores utilization-based settings.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      group = string

      # How this backend's capacity is measured: UTILIZATION (instance CPU,
      # the default — instance groups only), RATE (HTTP requests per second),
      # CONNECTION (open connections, for TCP/SSL), CUSTOM_METRICS
      # (backend-reported ORCA metrics), or IN_FLIGHT (concurrent in-flight
      # requests). IN_FLIGHT's target dial is not surfaced by the pinned
      # provider, so IN_FLIGHT backends ride the API's default target.
      # NEG backends must use RATE (or CUSTOM_METRICS); serverless NEGs
      # ignore balancing entirely. Mutable.
      balancing_mode = optional(string)

      # Fraction of the configured capacity this backend actually accepts
      # (default 1.0 = 100%). 0 drains the backend without removing it — the
      # standard lever for maintenance and gradual rollouts. Mutable.
      capacity_scaler = optional(number)

      # What this backend is (e.g. "blue pool, us-central1") — write it for
      # the operator reading a capacity page later. Mutable.
      description = optional(string, "")

      # Max simultaneous open connections for the whole backend (CONNECTION
      # mode target; optional ceiling in UTILIZATION mode). Mutable.
      max_connections = optional(number, 0)

      # Max simultaneous open connections per instance-group instance. Mutable.
      max_connections_per_instance = optional(number, 0)

      # Max simultaneous open connections per NEG endpoint. Mutable.
      max_connections_per_endpoint = optional(number, 0)

      # Max HTTP requests per second for the whole backend (RATE mode target;
      # optional ceiling in UTILIZATION mode). Mutable.
      max_rate = optional(number, 0)

      # Max HTTP requests per second per instance-group instance. Fractional
      # rates let small instances take partial shares. Mutable.
      max_rate_per_instance = optional(number, 0)

      # Max HTTP requests per second per NEG endpoint. Mutable.
      max_rate_per_endpoint = optional(number, 0)

      # Target CPU utilization (0.0-1.0) for UTILIZATION mode — the balancer
      # shifts new requests away as instances approach it. GCP's default is
      # 0.8. Ignored (and stripped by GCP) for NEG backends. Mutable.
      max_utilization = optional(number, 0)

      # Whether this backend is PREFERRED (filled to capacity before DEFAULT
      # backends receive traffic) — the primary/spillover pattern. Cannot be
      # set when the service's load_balancing_scheme is EXTERNAL. Global
      # backend services only (a regional passthrough service splits pools with
      # failover instead). Mutable.
      preference = optional(string, "")

      # Per-backend custom metrics for CUSTOM_METRICS balancing mode, reported
      # by this backend via ORCA. Each can run dry (reported but not acted on)
      # while being validated.
      custom_metrics = optional(list(object({
        # Metric name as reported by the backend in ORCA load reports (e.g. a
        # named utilization gauge). Must match what the backend actually emits.
        name = string

        # Report the metric without acting on it — the safe first step while
        # validating that backends emit sane values.
        dry_run = optional(bool, false)

        # Target utilization (0.0-1.0) for this metric, above which the balancer
        # shifts new requests away. GCP's default is 0.8.
        max_utilization = optional(number)
      })), [])

      # Mark this backend as part of the FAILOVER pool of a passthrough Network
      # Load Balancer: it receives traffic only when the primary pool's healthy
      # ratio drops to failover_policy.failover_ratio (or every primary backend
      # is unhealthy). Several backends may be failover backends. Regional
      # backend services only. Mutable.
      failover = optional(bool, false)
    })), [])

    # How requests from the same client stick to the same backend (default
    # NONE — every request is balanced independently). Cookie-based modes
    # (GENERATED_COOKIE, HTTP_COOKIE, STRONG_COOKIE_AFFINITY) need an
    # HTTP-family protocol; CLIENT_IP modes hash on network attributes.
    # CLIENT_IP_NO_DESTINATION — regional services only — hashes on the
    # client IP alone, the mode for an internal passthrough Network Load
    # Balancer used as a next hop. Session affinity is best-effort, not a
    # guarantee — backends going unhealthy still break affinity. Not
    # applicable when protocol is UDP; not allowed together with ha_policy.
    # Mutable.
    session_affinity = optional(string)

    # Lifetime in seconds of the cookie GCP generates for GENERATED_COOKIE
    # session affinity (0, the default, makes it a non-persistent session
    # cookie; max 86400 = 1 day). Only meaningful with GENERATED_COOKIE.
    # Mutable.
    affinity_cookie_ttl_sec = optional(number, 0)

    # The cookie GCP uses for STRONG_COOKIE_AFFINITY — stronger stickiness
    # than GENERATED_COOKIE because the cookie encodes the exact backend
    # endpoint. Only valid with session_affinity STRONG_COOKIE_AFFINITY, and
    # optional there: choosing that mode is the whole statement, and the
    # modules send GCP the cookie configuration it requires (GCP's generated
    # cookie name, whole-site path, session lifetime) when this block is
    # absent. Declare it only to customize the cookie.
    strong_session_affinity_cookie = optional(object({
      # Cookie name the load balancer sets and matches. Empty uses GCP's
      # generated default name.
      name = optional(string, "")

      # Path attribute of the cookie — limit affinity to a URL subtree (e.g.
      # /app). Empty applies to the whole site.
      path = optional(string, "")

      # Cookie lifetime. Zero/unset makes it a non-persistent session cookie
      # that vanishes when the browser closes.
      ttl = optional(object({
        # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
        seconds = optional(number, 0)

        # Fraction of a second at nanosecond resolution (0 to 999,999,999).
        # Durations under one second use seconds = 0 and a positive nanos.
        nanos = optional(number, 0)
      }))
    }))

    # The load balancing algorithm used within each backend group once the
    # group is chosen (GCP default ROUND_ROBIN). LEAST_REQUEST and the
    # hash-based policies (RING_HASH, MAGLEV) matter for uneven request
    # costs and soft session affinity; WEIGHTED_ROUND_ROBIN balances on
    # backend-reported custom metrics. Only ROUND_ROBIN and RING_HASH are
    # supported for proxyless gRPC. Mutable.
    locality_lb_policy = optional(string, "")

    # Ordered list of locality LB policies for Traffic Director deployments
    # that need a custom (xDS-configured) policy with built-in fallbacks.
    # Each entry is either a built-in policy name or a custom policy plus its
    # opaque configuration; Traffic Director uses the first one it supports.
    # Overrides locality_lb_policy when set. Mutable.
    locality_lb_policies = optional(list(object({
      # A built-in locality policy by name. The WEIGHTED_* policies are not
      # valid inside this list — use the top-level locality_lb_policy for
      # those.
      policy = optional(object({
        # The built-in policy name.
        name = string
      }))

      # A custom policy implemented in the xDS client (Envoy/gRPC), selected
      # by name with an opaque configuration string. Traffic Director falls
      # back to the next entry if the client does not recognize it.
      custom_policy = optional(object({
        # Identifier of the custom policy as registered in the xDS client (e.g.
        # an Envoy load balancing extension name).
        name = string

        # Opaque configuration handed to the custom policy, in whatever format
        # the policy implementation expects (commonly JSON).
        data = optional(string, "")
      }))
    })), [])

    # Parameters for consistent-hash load balancing — soft session affinity
    # where a backend's share of the hash ring survives other backends
    # joining or leaving. Only applies with load_balancing_scheme
    # INTERNAL_SELF_MANAGED and locality_lb_policy MAGLEV or RING_HASH.
    consistent_hash = optional(object({
      # Hash on an HTTP cookie, generating it when absent — soft session
      # affinity for clients that keep cookies. Only applies when
      # session_affinity is HTTP_COOKIE.
      http_cookie = optional(object({
        # Cookie name to hash on (and to generate when absent).
        name = optional(string, "")

        # Path attribute set when the cookie is generated.
        path = optional(string, "")

        # Lifetime of the generated cookie. Zero/unset makes it a session
        # cookie.
        ttl = optional(object({
          # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
          seconds = optional(number, 0)

          # Fraction of a second at nanosecond resolution (0 to 999,999,999).
          # Durations under one second use seconds = 0 and a positive nanos.
          nanos = optional(number, 0)
        }))
      }))

      # Hash on the value of this request header. Only applies when
      # session_affinity is HEADER_FIELD.
      http_header_name = optional(string, "")

      # Minimum number of virtual nodes on the hash ring (default 1024).
      # Larger rings spread load more evenly across backends at slightly
      # higher memory cost; must be at least the number of backend hosts.
      minimum_ring_size = optional(number)
    }))

    # Cache responses at Google's edge with Cloud CDN. Off by default:
    # without it every request is proxied to a backend. Only valid on
    # external schemes (EXTERNAL, EXTERNAL_MANAGED) — Cloud CDN does not
    # front internal load balancers. Turning it on activates cdn_policy (or
    # sensible CDN defaults when cdn_policy is omitted). Mutable.
    enable_cdn = optional(bool, false)

    # How Cloud CDN caches responses from these backends. Only meaningful
    # with enable_cdn — GCP ignores the policy while CDN is off.
    cdn_policy = optional(object({
      # What gets cached. CACHE_ALL_STATIC (the GCP default) caches static
      # content types and honors origin cache headers for the rest;
      # USE_ORIGIN_HEADERS caches only what the backends explicitly mark
      # cacheable (TTL fields must be unset — the origin controls lifetimes);
      # FORCE_CACHE_ALL caches everything, ignoring origin headers (never
      # combine with private or per-user content; max_ttl must be unset).
      cache_mode = optional(string, "")

      # Seconds a response may be cached by browsers and other downstream
      # caches (sets the max-age clients see; GCP default 3600, max 86400).
      # Keep it shorter than default_ttl so edge caches revalidate before
      # clients do.
      client_ttl = optional(number, 0)

      # Seconds the edge caches a response when the origin sets no caching
      # headers (GCP default 3600, max 31622400 = 1 year). The workhorse TTL
      # for CACHE_ALL_STATIC and FORCE_CACHE_ALL.
      default_ttl = optional(number, 0)

      # Upper bound in seconds on any cache lifetime, capping even origin
      # headers that ask for longer (GCP default 86400, max 31622400). Not
      # allowed with USE_ORIGIN_HEADERS or FORCE_CACHE_ALL cache modes.
      max_ttl = optional(number, 0)

      # Cache error responses (404s, redirects) at the edge so failing paths
      # do not hammer the backends. Pair with negative_caching_policy to set
      # per-status TTLs; without it GCP applies default lifetimes.
      negative_caching = optional(bool, false)

      # Per-status-code TTLs for negative caching. Only effective with
      # negative_caching enabled. Codes limited by GCP to 300, 301, 308, 404,
      # 405, 410, 421, 451, and 501.
      negative_caching_policy = optional(list(object({
        # The HTTP status code to cache. GCP supports 300, 301, 308, 404, 405,
        # 410, 421, 451, and 501.
        code = number

        # Seconds responses with this status are cached at the edge
        # (0 to 1800 = 30 minutes).
        ttl = optional(number, 0)
      })), [])

      # Seconds the edge may keep serving a stale response while it
      # revalidates with the origin in the background (max 86400; 0 disables).
      # Smooths over brief backend outages for content that tolerates slight
      # staleness.
      serve_while_stale = optional(number, 0)

      # Collapse concurrent cache-miss requests for the same object into one
      # origin fetch. Protects the backends from thundering herds on cache
      # expiry of popular objects.
      request_coalescing = optional(bool, false)

      # Seconds a response to a SIGNED request stays fresh in the cache before
      # revalidation (GCP default 3600, max 86400). Only meaningful with
      # signed URLs or cookies; the signature's own expiry still governs
      # access.
      signed_url_cache_max_age_sec = optional(number)

      # What forms the cache key beyond the URL. The backend-service flavor is
      # richer than a backend bucket's: host, protocol, query handling, named
      # cookies, and headers can all join or leave the key. Leave unset for
      # GCP's default (host + protocol + full query string).
      cache_key_policy = optional(object({
        # Include the request host in the cache key. GCP's default is true —
        # turn it off only when several hosts genuinely serve identical content.
        include_host = optional(bool, false)

        # Include the protocol (http/https) in the cache key. GCP's default is
        # true — turn it off only when both schemes serve identical bytes.
        include_protocol = optional(bool, false)

        # Include the query string in the cache key. GCP's default is true.
        # When true, narrow it with query_string_whitelist or blacklist; when
        # false, the query string is ignored entirely (and the lists must be
        # unset).
        include_query_string = optional(bool, false)

        # Query parameters included in the cache key, all others ignored.
        # Include only parameters that genuinely change the response so
        # equivalent requests share a cache entry. Mutually exclusive with
        # query_string_blacklist.
        query_string_whitelist = optional(list(string), [])

        # Query parameters excluded from the cache key, all others included —
        # for stripping tracking parameters (utm_*) that never change the
        # response. Mutually exclusive with query_string_whitelist.
        query_string_blacklist = optional(list(string), [])

        # Request headers whose values join the cache key — for backends that
        # vary responses by header (e.g. Accept for image format negotiation).
        # Each distinct value creates a separate cache entry, so keep this list
        # short.
        include_http_headers = optional(list(string), [])

        # Cookie names whose values join the cache key — for backends that vary
        # cached content by cookie (e.g. an A/B bucket cookie). Each distinct
        # value creates a separate cache entry.
        include_named_cookies = optional(list(string), [])
      }))

      # Skip the cache entirely for requests carrying any of these headers
      # (at most 5) — an escape hatch for debugging or per-request freshness
      # (e.g. a Pragma: no-cache internal tooling header).
      bypass_cache_on_request_headers = optional(list(object({
        # The header name to match (case-insensitive); any value triggers the
        # bypass.
        header_name = string
      })), [])
    }))

    # Cloud Armor security policy evaluated on every request AFTER the CDN
    # cache (protects the backends: WAF rules, rate limiting, geo/IP
    # blocking). Reference a GcpCloudArmorPolicy of type CLOUD_ARMOR —
    # edge policies are not valid here. The scopes must match: a regional
    # backend service attaches only a regional Cloud Armor policy in its own
    # region, a global one only a global policy. Mutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_policy = optional(string, "")

    # Cloud Armor EDGE security policy filtering requests BEFORE the CDN
    # cache (protects cached content: geo/IP blocking at the edge).
    # Reference a GcpCloudArmorPolicy of type CLOUD_ARMOR_EDGE — standard
    # backend policies are not valid here. Global backend services only (the
    # edge is Google's global CDN). Mutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    edge_security_policy = optional(string, "")

    # Identity-Aware Proxy: authenticate every request against Google
    # identities before it reaches the backends — zero-trust access to
    # internal tools without a VPN. Requests arrive with IAP assertion
    # headers the backend can trust. HTTPS frontends only.
    iap = optional(object({
      # Turn IAP enforcement on. When enabled, every request must carry a
      # valid Google identity; unauthenticated requests get a login redirect.
      enabled = optional(bool, false)

      # OAuth2 client ID of a custom IAP client. Leave both id and secret
      # empty to use the Google-managed OAuth client.
      oauth2_client_id = optional(string, "")

      # OAuth2 client secret paired with oauth2_client_id. Handled as a
      # secret: never stored in plaintext in the control plane, never exposed
      # in outputs (GCP itself only ever returns its SHA-256 after creation).
      oauth2_client_secret = optional(string, "")
    }))

    # Request logging to Cloud Logging for this backend service. Off by
    # default. Sampling keeps log volume (and cost) proportional on
    # high-traffic services.
    log_config = optional(object({
      # Write request logs to Cloud Logging. Off by default.
      enable = optional(bool, false)

      # Fraction of requests logged, 0.0-1.0 (GCP default 1.0 = everything).
      # Sample aggressively on high-QPS services — full logging is a real
      # cost line.
      sample_rate = optional(number)

      # Which optional fields join each log entry: INCLUDE_ALL_OPTIONAL,
      # EXCLUDE_ALL_OPTIONAL (the GCP default), or CUSTOM (name them in
      # optional_fields).
      optional_mode = optional(string, "")

      # Names of the optional log fields to include with optional_mode CUSTOM
      # (e.g. tls.protocol, orca_load_report).
      optional_fields = optional(list(string), [])

      # HTTP request headers whose values join each log entry (e.g.
      # "X-Request-Id", "User-Agent") — for tracing a request across services
      # without instrumenting the backend. Requires enable and an HTTP-family
      # protocol (HTTP, HTTPS, HTTP2, GRPC). Header names are case-insensitive
      # in HTTP; each entry is one name.
      request_headers = optional(list(string), [])

      # HTTP response headers whose values join each log entry (e.g.
      # "Content-Type", a backend's own "X-Cache" or "X-Served-By"). Same
      # preconditions as request_headers.
      response_headers = optional(list(string), [])
    }))

    # Failover behavior for a passthrough Network Load Balancer whose
    # backends are split into primary and failover pools (backends[].failover
    # marks the failover pool): when to shift to the failover pool, whether
    # to drop traffic if both pools are unhealthy, and whether to drain
    # existing connections on failover. Regional backend services only; not
    # allowed together with ha_policy.
    failover_policy = optional(object({
      # Skip connection draining when traffic fails over (or back): existing
      # connections to the old active pool are cut rather than drained for the
      # fixed 10-minute window. TCP only. GCP default false.
      disable_connection_drain_on_failover = optional(bool)

      # When NO backend in either pool is healthy, drop new connections (true)
      # instead of spraying them across every primary backend in the hope one
      # answers (false, GCP's default).
      drop_traffic_if_unhealthy = optional(bool)

      # The healthy ratio (0.0-1.0) of the primary pool at or below which
      # traffic moves to the failover pool. Unset means traffic fails over only
      # when every primary backend is unhealthy. When the failover pool is
      # itself all-unhealthy, traffic returns to the primary pool best-effort.
      failover_ratio = optional(number)
    }))

    # How a passthrough Network Load Balancer tracks connections for session
    # consistency: per connection or per session, whether tracked flows
    # persist to a backend that turned unhealthy, and how long idle entries
    # live. Regional backend services only; not allowed together with
    # ha_policy.
    connection_tracking_policy = optional(object({
      # What identifies a tracked flow: PER_CONNECTION (the GCP default) keys on
      # the protocol's full connection tuple; PER_SESSION keys on the configured
      # session_affinity, so a client's whole session sticks together. Both
      # engines send the default explicitly when this is empty.
      tracking_mode = optional(string)

      # What happens to tracked flows when their backend turns unhealthy:
      # DEFAULT_FOR_PROTOCOL (the GCP default) keeps TCP/SCTP connections on the
      # unhealthy backend when tracking is per-connection or 5-tuple affinity,
      # never UDP; NEVER_PERSIST always diverts them to healthy backends;
      # ALWAYS_PERSIST keeps them where they are. Both engines send the default
      # explicitly when this is empty.
      connection_persistence_on_unhealthy_backends = optional(string)

      # Seconds a connection-tracking entry lives with no matching traffic. For
      # the internal passthrough NLB the minimum (and GCP default) is 600 and
      # the maximum 57600; for the external passthrough NLB it must be 60 when
      # tracking per session with CLIENT_IP or CLIENT_IP_PROTO affinity, and
      # the default otherwise. Left unset, GCP computes the default and the
      # engines send nothing (the argument is computed by the API), so an
      # untouched value never shows as drift.
      idle_timeout_sec = optional(number)

      # Strong session affinity for the external passthrough Network Load
      # Balancer: track flows so a session keeps its backend across connection
      # churn. Google documents this option as not yet publicly available; it
      # is here so the spec matches the provider surface. Default false.
      enable_strong_affinity = optional(bool, false)
    }))

    # High-availability IP failover for an internal (or external) passthrough
    # Network Load Balancer with exactly one leader backend at a time: a
    # single backend group (or one endpoint inside it) holds the VIP, and
    # fast_ip_move lets the VIP move with a gratuitous ARP / router
    # advertisement instead of waiting on health checks. Regional backend
    # services only. Google forbids it together with health_check,
    # session_affinity, failover_policy, and connection_tracking_policy — the
    # leader IS the routing decision.
    ha_policy = optional(object({
      # How the VIP moves to a new leader: DISABLED (the leader changes only
      # through the haPolicy.leader API, i.e. by editing leader below) or
      # GARP_RA (the VM that should become leader announces itself with a
      # gratuitous ARP for IPv4 or a Router Advertisement for IPv6 and Google
      # moves the VIP within seconds — the mechanism for keepalived-style
      # active/passive pairs). Immutable: changing it recreates the backend
      # service.
      fast_ip_move = optional(string, "")

      # The current leader: the zonal network endpoint group holding the VIP
      # and, optionally, the exact instance inside it. Mutable — editing this
      # is the API-driven leader change.
      leader = optional(object({
        # Fully-qualified URL of the zonal network endpoint group the leader is
        # attached to. Must be one of this service's backends.
        backend_group = optional(string, "")

        # The leader endpoint inside that group.
        network_endpoint = optional(object({
          # Name of the VM instance serving as the leader. The instance must
          # already be attached to the leader's backend_group NEG.
          instance = optional(string, "")
        }))
      }))
    }))

    # Zonal affinity for a passthrough Network Load Balancer: keep traffic
    # inside the client's zone and decide whether it may spill to other zones
    # when the local zone's healthy capacity drops below a ratio. Regional
    # backend services only.
    network_pass_through_lb_traffic_policy = optional(object({
      # Keep new connections inside the client's zone while that zone has
      # enough healthy backends, spilling to other zones only below a ratio.
      zonal_affinity = optional(object({
        # The mode: ZONAL_AFFINITY_DISABLED (the GCP default — connections spread
        # across all zones), ZONAL_AFFINITY_SPILL_CROSS_ZONE (stay in the
        # client's zone while its healthy ratio is at or above spillover_ratio,
        # otherwise use every zone), or ZONAL_AFFINITY_STAY_WITHIN_ZONE (never
        # leave the zone, even when it has no healthy backend). Both engines send
        # the default explicitly when this is empty.
        spillover = optional(string)

        # The healthy ratio (0.0-1.0) of the client's zone at or above which new
        # connections stay local; below it they spread across all zones (SPILL
        # mode only). Unset lets GCP apply its default.
        spillover_ratio = optional(number)
      }))
    }))

    # Headers the load balancer ADDS to requests before forwarding them to
    # the backends, in "Header-Name: value" form. Values may use variables
    # like {client_ip} or {tls_version}. Typical uses: passing the client's
    # geo data or TLS parameters to the application. Global backend services
    # only. Mutable.
    custom_request_headers = optional(list(string), [])

    # Headers the load balancer ADDS to responses before returning them to
    # clients, in "Header-Name: value" form. Values may use variables like
    # {cdn_cache_status}. Typical uses: security headers
    # (Strict-Transport-Security) and cache observability. Global backend
    # services only. Mutable.
    custom_response_headers = optional(list(string), [])

    # Whether the load balancer compresses responses (gzip/brotli) for
    # clients that ask for it. AUTOMATIC compresses compressible content
    # types; DISABLED (the GCP default when unset) never compresses.
    # Compression is applied by the load balancer — backends keep serving
    # uncompressed responses. Global backend services only. Mutable.
    compression_mode = optional(string, "")

    # Connection-volume limits protecting backends from overload — the
    # service-mesh circuit breaker. Only applies with load_balancing_scheme
    # INTERNAL_SELF_MANAGED (Traffic Director).
    circuit_breakers = optional(object({
      # Max concurrent connections to the whole backend service (GCP default
      # 1024).
      max_connections = optional(number)

      # Max requests queued waiting for a connection (GCP default 1024).
      max_pending_requests = optional(number)

      # Max concurrent requests to the whole backend service (GCP default
      # 1024).
      max_requests = optional(number)

      # Max requests per connection — 1 disables HTTP keep-alive. Unset means
      # unlimited.
      max_requests_per_connection = optional(number, 0)

      # Max concurrent retries across the backend service (GCP default 3).
      # Retries amplify load during incidents — keep this bounded.
      max_retries = optional(number)
    }))

    # Passive health checking: eject backends that keep erroring from the
    # load balancing pool for a cooling-off period, without waiting for the
    # active health check to fail. Only applies with load_balancing_scheme
    # INTERNAL_SELF_MANAGED or EXTERNAL_MANAGED.
    outlier_detection = optional(object({
      # Base duration a host stays ejected; actual ejection time is this
      # multiplied by the number of times the host has been ejected (GCP
      # default 30s).
      base_ejection_time = optional(object({
        # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
        seconds = optional(number, 0)

        # Fraction of a second at nanosecond resolution (0 to 999,999,999).
        # Durations under one second use seconds = 0 and a positive nanos.
        nanos = optional(number, 0)
      }))

      # Consecutive 5xx responses (or connection errors) before ejection (GCP
      # default 5).
      consecutive_errors = optional(number, 0)

      # Consecutive gateway-class failures (502/503/504) before ejection (GCP
      # default 5). Catches infrastructure failures faster than
      # consecutive_errors on mixed error streams.
      consecutive_gateway_failure = optional(number, 0)

      # Percentage chance (0-100) that a host is ACTUALLY ejected when
      # consecutive_errors trips (GCP default 100). Lower values ease the
      # policy in gradually.
      enforcing_consecutive_errors = optional(number, 0)

      # Percentage chance (0-100) of ejection when consecutive_gateway_failure
      # trips (GCP default 0 — off unless raised).
      enforcing_consecutive_gateway_failure = optional(number, 0)

      # Percentage chance (0-100) of ejection when a host's success rate falls
      # statistically below the pool (GCP default 100).
      enforcing_success_rate = optional(number, 0)

      # How often ejection sweeps run (GCP default 1s).
      interval = optional(object({
        # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
        seconds = optional(number, 0)

        # Fraction of a second at nanosecond resolution (0 to 999,999,999).
        # Durations under one second use seconds = 0 and a positive nanos.
        nanos = optional(number, 0)
      }))

      # Max percentage (0-100) of the pool that may be ejected at once (GCP
      # default 10) — the safety valve that keeps outlier detection from
      # draining the whole service.
      max_ejection_percent = optional(number, 0)

      # Minimum number of hosts in the pool before success-rate ejection
      # activates (GCP default 5) — below it the statistics are meaningless.
      success_rate_minimum_hosts = optional(number, 0)

      # Minimum requests a host must have received in the interval for its
      # success rate to count (GCP default 100).
      success_rate_request_volume = optional(number, 0)

      # How many standard deviations below the pool mean a host's success
      # rate must fall to be ejected, multiplied by 1000 (GCP default 1900 =
      # 1.9 stdev). Lower is more aggressive.
      success_rate_stdev_factor = optional(number, 0)
    }))

    # Default maximum duration for streams to this service, computed from
    # stream start until the response is completely processed (including
    # retries). Unset means no timeout limit. Can be overridden per-route in
    # the URL map. Only allowed with load_balancing_scheme
    # INTERNAL_SELF_MANAGED.
    max_stream_duration = optional(object({
      # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
      seconds = optional(number, 0)

      # Fraction of a second at nanosecond resolution (0 to 999,999,999).
      # Durations under one second use seconds = 0 and a positive nanos.
      nanos = optional(number, 0)
    }))

    # Backend authentication and TLS settings for Traffic Director
    # (client TLS policy, SAN validation) and for AWS-hosted internet-NEG
    # origins (Signature Version 4 request signing).
    security_settings = optional(object({
      # Self-link of a networksecurity ClientTlsPolicy describing how the
      # load balancer authenticates itself to the backends (mTLS). Traffic
      # Director only. Plain URL — the policy is a Network Security resource
      # outside the compute family.
      client_tls_policy = optional(string, "")

      # Subject Alternative Names the backend's server certificate must
      # present — pins backend identity for Traffic Director mTLS.
      subject_alt_names = optional(list(string), [])

      # Sign origin requests with AWS Signature Version 4 — for internet-NEG
      # backends fronting private S3 buckets or other SigV4-authenticated AWS
      # origins.
      aws_v4_authentication = optional(object({
        # AWS access key ID — the username-like identifier of the key pair (not
        # itself a secret).
        access_key_id = optional(string, "")

        # AWS secret access key paired with access_key_id. Handled as a secret:
        # never stored in plaintext in the control plane, and GCP never returns
        # it on reads.
        access_key = optional(string, "")

        # Optional version identifier for the key, echoed in logs to trace
        # which credential signed a request during rotation.
        access_key_version = optional(string, "")

        # AWS region of the origin (e.g. us-east-1) — part of the SigV4 signing
        # scope.
        origin_region = optional(string, "")
      }))
    }))

    # TLS parameters for the load balancer's connections TO the backends:
    # which server certificate authentication to apply and what SNI to send.
    # Only valid when protocol is SSL, HTTPS, or HTTP2.
    tls_settings = optional(object({
      # Self-link of a networksecurity BackendAuthenticationConfig that
      # validates the backend's server certificate (trust anchor + client
      # cert). Plain URL — the config is a Network Security resource outside
      # the compute family.
      authentication_config = optional(string, "")

      # Server Name Indication sent in the TLS handshake to the backends —
      # for origins that route or select certificates by SNI.
      sni = optional(string, "")

      # Subject Alternative Names the backend certificate must match, each a
      # DNS name or a URI. GCP allows at most 5.
      subject_alt_names = optional(list(object({
        # A DNS-name SAN (e.g. origin.example.com).
        dns_name = optional(string)

        # A URI SAN (e.g. spiffe://cluster/ns/prod/sa/web).
        uniform_resource_identifier = optional(string)
      })), [])
    }))

    # Whether the load balancer prefers IPv4 or IPv6 addresses when
    # connecting to dual-stack backends. Unset uses GCP's default (IPv4).
    # Mutable.
    ip_address_selection_policy = optional(string, "")

    # Canary state for migrating this backend service from the classic
    # EXTERNAL scheme to EXTERNAL_MANAGED without recreating it: PREPARE
    # first, then optionally TEST_BY_PERCENTAGE, then TEST_ALL_TRAFFIC —
    # after which load_balancing_scheme can be flipped to EXTERNAL_MANAGED.
    # Only meaningful while the scheme is still EXTERNAL. Mutable.
    external_managed_migration_state = optional(string, "")

    # Fraction of traffic (0-100) sent to the envoy-based global external
    # ALB during a TEST_BY_PERCENTAGE canary migration. Only meaningful with
    # external_managed_migration_state TEST_BY_PERCENTAGE. Mutable.
    external_managed_migration_testing_percentage = optional(number, 0)

    # Custom metrics the WEIGHTED_ROUND_ROBIN locality policy balances on,
    # reported by the backends via the Open Request Cost Aggregation (ORCA)
    # protocol. Only meaningful with locality_lb_policy
    # WEIGHTED_ROUND_ROBIN.
    custom_metrics = optional(list(object({
      # Metric name as reported by the backends in ORCA load reports.
      name = string

      # Report the metric without acting on it — the safe first step while
      # validating that backends emit sane values.
      dry_run = optional(bool, false)
    })), [])

    # Self-link of a networkservices ServiceLbPolicy attaching advanced
    # traffic-distribution features (e.g. auto-capacity failover) to this
    # backend service. Plain URL — the service LB policy is a Network
    # Services resource outside the compute family. Global backend services
    # only. Mutable.
    service_lb_policy = optional(string, "")

    # Keys for signing Cloud CDN signed URLs and signed cookies — the
    # mechanism for serving private content from the cache with expiring,
    # tamper-proof links. GCP allows at most 3 keys per backend service so
    # one can be rotated while another stays live. Each key's material is a
    # secret; rotate by adding a new key, re-signing URLs, then removing the
    # old one.
    signed_url_keys = optional(list(object({
      # Name of the key, referenced by the key_name parameter of signed URLs.
      # Must be 1-63 characters: lowercase letters, digits, and hyphens; must
      # start with a letter and end with a letter or digit. Immutable:
      # renaming replaces the key, invalidating URLs signed with the old name.
      name = string

      # The 128-bit signing key, base64url-encoded (RFC 4648 §5) — generate
      # one with: head -c 16 /dev/urandom | base64 | tr '+/' '-_'. 22
      # characters of base64url, with or without the trailing == padding.
      # Anyone holding this value can mint valid signed URLs, so it is
      # handled as a secret. Immutable per key name: rotating means adding a
      # new key and removing the old. The base64url shape is taught here
      # rather than enforced by a validation rule, because sensitive fields
      # hold a managed-secret reference on consuming platforms and a
      # content-shape rule would reject every reference.
      key_value = string
    })), [])

    # Resource Manager tags bound to the backend service for org-policy and
    # IAM conditions. Keys in the form "tagKeys/{id}", values
    # "tagValues/{id}". Create-time only: changing them later replaces the
    # backend service.
    resource_manager_tags = optional(map(string), {})

    # Deletion policy for the backend service AND its signed-URL keys — one
    # switch governs both objects this kind manages:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- both are deleted (GCP refuses to delete the backend
    #                service while a URL map or forwarding rule still
    #                references it); the backends it pointed at — instance
    #                groups, NEGs — are untouched
    #   "PREVENT" -- destroy FAILS; protects the routing target of a live
    #                load balancer
    #   "ABANDON" -- both are removed from management but keep serving in GCP
    deletion_policy = optional(string, "")
  })
}
