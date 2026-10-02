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
  description = "GcpBackendBucket specification"
  type = object({
    # The GCP project that owns the backend bucket (which may differ from the
    # project owning the GCS bucket — cross-project origins are valid).
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable: changing it destroys and recreates the backend bucket.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the backend bucket in GCP. Must be 1-63 characters: lowercase
    # letters, digits, and hyphens; must start with a letter and end with a
    # letter or digit. If not specified, defaults to metadata.name.
    # Immutable: changing it destroys and recreates the backend bucket,
    # briefly breaking every URL map that references the old self_link.
    backend_bucket_name = optional(string, "")

    # The Cloud Storage bucket whose objects are served — the origin.
    # Reference a GcpGcsBucket resource or provide the bucket name directly.
    # Mutable: pointing at a different bucket is an in-place update, which
    # makes origin swaps (e.g. blue/green static releases) cheap.
    # Objects must be publicly readable (or served via signed URLs/cookies) —
    # the load balancer does not authenticate to the bucket.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    bucket_name = string

    # What this backend bucket serves and which URL maps use it — write it for
    # the operator tracing a route later. Mutable.
    description = optional(string, "")

    # Cache responses at Google's edge with Cloud CDN. Off by default: without
    # it every request is proxied to the bucket. Turning it on activates
    # cdn_policy (or sensible CDN defaults when cdn_policy is omitted).
    # Cannot be enabled together with load_balancing_scheme INTERNAL_MANAGED —
    # Cloud CDN only fronts external load balancers. Mutable.
    enable_cdn = optional(bool, false)

    # How Cloud CDN caches responses from this origin. Only meaningful with
    # enable_cdn — GCP ignores the policy while CDN is off.
    cdn_policy = optional(object({
      # What gets cached. CACHE_ALL_STATIC (the GCP default) caches static
      # content types and honors origin cache headers for the rest;
      # USE_ORIGIN_HEADERS caches only what the origin explicitly marks
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

      # Cache error responses (404s, redirects) at the edge so failing paths do
      # not hammer the origin. Pair with negative_caching_policy to set
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

      # Seconds the edge may keep serving a stale response while it revalidates
      # with the origin in the background (max 86400; 0 disables). Smooths over
      # brief origin outages for content that tolerates slight staleness.
      serve_while_stale = optional(number, 0)

      # Collapse concurrent cache-miss requests for the same object into one
      # origin fetch. Protects the origin from thundering herds on cache
      # expiry of popular objects.
      request_coalescing = optional(bool, false)

      # Seconds a response to a SIGNED request stays fresh in the cache before
      # revalidation (max 86400). Only meaningful with signed URLs or cookies;
      # after this window the edge revalidates, though the signature's own
      # expiry still governs access.
      signed_url_cache_max_age_sec = optional(number, 0)

      # What forms the cache key beyond the URL host and path. Leave unset to
      # ignore query strings and headers entirely — the best hit rate for
      # immutable, fingerprinted assets.
      cache_key_policy = optional(object({
        # Query parameters included in the cache key (all others are ignored).
        # Include only parameters that genuinely change the response — e.g. an
        # image resizer's "w" and "h" — so equivalent requests share a cache
        # entry.
        query_string_whitelist = optional(list(string), [])

        # Request headers whose values join the cache key — for origins that vary
        # responses by header (e.g. Accept for image format negotiation). Each
        # distinct value creates a separate cache entry, so keep this list short.
        include_http_headers = optional(list(string), [])
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

    # Whether the load balancer compresses responses (gzip/brotli) for clients
    # that ask for it. AUTOMATIC compresses compressible content types;
    # DISABLED (the GCP default when unset) never compresses. Compression is
    # applied by the load balancer, not the bucket — objects stay uncompressed
    # at the origin. Mutable.
    compression_mode = optional(string, "")

    # Response headers the load balancer adds to every response served from
    # this backend, in "Header-Name: value" form. Values may use variables
    # like {cdn_cache_status}. Typical uses: security headers
    # (Strict-Transport-Security) and cache observability
    # (X-Cache-Status: {cdn_cache_status}). Mutable.
    custom_response_headers = optional(list(string), [])

    # Cloud Armor EDGE security policy filtering requests before they reach
    # the cache or the origin (rate limiting and geo/IP blocking at the edge).
    # Reference a GcpCloudArmorPolicy of type CLOUD_ARMOR_EDGE — standard
    # backend policies are not valid here. Mutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    edge_security_policy = optional(string, "")

    # Which load balancer family this backend serves. Leave unset for global
    # EXTERNAL HTTP(S) load balancers — the overwhelmingly common case for
    # static content. INTERNAL_MANAGED serves cross-region internal
    # Application Load Balancers instead, and is incompatible with Cloud CDN.
    # Immutable: changing it destroys and recreates the backend bucket.
    load_balancing_scheme = optional(string, "")

    # Keys for signing Cloud CDN signed URLs and signed cookies — the
    # mechanism for serving private content from the cache with expiring,
    # tamper-proof links. GCP allows at most 3 keys per backend bucket so one
    # can be rotated while another stays live. Each key's material is a
    # secret; rotate by adding a new key, re-signing URLs, then removing the
    # old one.
    signed_url_keys = optional(list(object({
      # Name of the key, referenced by the key_name parameter of signed URLs.
      # Must be 1-63 characters: lowercase letters, digits, and hyphens; must
      # start with a letter and end with a letter or digit. Immutable: renaming
      # replaces the key, invalidating URLs signed with the old name.
      name = string

      # The 128-bit signing key, base64url-encoded (RFC 4648 §5) — generate one
      # with: head -c 16 /dev/urandom | base64 | tr '+/' '-_'. 22 characters of
      # base64url, with or without the trailing == padding. Anyone holding this
      # value can mint valid signed URLs, so it is handled as a secret.
      # Immutable per key name: rotating means adding a new key and removing
      # the old. The base64url shape is taught here rather than enforced by a
      # validation rule, because sensitive fields hold a managed-secret
      # reference on consuming platforms and a content-shape rule would
      # reject every reference.
      key_value = string
    })), [])

    # Resource Manager tags bound to the backend bucket for org-policy and
    # IAM conditions. Keys in the form "tagKeys/{id}", values
    # "tagValues/{id}". Create-time only: changing them later replaces the
    # backend bucket.
    resource_manager_tags = optional(map(string), {})

    # Deletion policy for the backend bucket AND its signed-URL keys — one
    # switch governs both objects this kind manages:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- both are deleted (GCP refuses to delete the backend
    #                bucket while a URL map still references it); the GCS
    #                bucket behind it is untouched — it belongs to its own
    #                kind
    #   "PREVENT" -- destroy FAILS; protects a CDN origin that URL maps may
    #                still route to
    #   "ABANDON" -- both are removed from management but keep serving in GCP
    deletion_policy = optional(string, "")
  })
}
