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
  description = "AwsCloudFront specification"
  type = object({
    # The AWS region used for the provider connection. CloudFront itself
    # is a GLOBAL service -- the distribution is the same everywhere and
    # its supporting resources (certificates, CLOUDFRONT-scope WAF ACLs)
    # must live in us-east-1 -- but the deployment still runs through a
    # regional endpoint. Example: "us-east-1".
    region = string

    # Whether the distribution accepts and serves viewer requests. True
    # (the default) serves traffic; false keeps the distribution
    # deployed-but-dark (useful for staging a configuration before
    # cutover, and required by AWS before a distribution can be
    # deleted -- the modules handle that disable-on-destroy dance).
    enabled = optional(bool)

    # Alternate domain names (CNAMEs) the distribution answers for,
    # e.g. "cdn.example.com". Requires viewer_certificate with your own
    # certificate covering every alias -- CloudFront rejects an alias
    # the certificate does not cover.
    aliases = optional(list(string), [])

    # Free-form comment shown in the AWS Console distribution list.
    # Up to 128 characters.
    comment = optional(string, "")

    # The object CloudFront serves when a viewer requests the root URL
    # ("/"), e.g. "index.html". Applies to the root only -- subdirectory
    # index documents need a CloudFront Function or S3 website origin.
    default_root_object = optional(string, "")

    # The maximum HTTP version viewers can use: "http2" (the default),
    # "http1.1", "http2and3", or "http3". HTTP/3 (QUIC) improves
    # performance on lossy networks at no extra cost -- "http2and3" is
    # the safe way to adopt it. Empty keeps http2.
    http_version = optional(string, "")

    # Whether CloudFront answers IPv6 viewer requests. False (the AWS
    # default) is IPv4-only; enabling costs nothing and serves
    # dual-stack (create the AAAA alias record alongside the A record).
    is_ipv6_enabled = optional(bool, false)

    # Which edge locations serve the distribution -- the cost/latency
    # dial: "PriceClass_All" (every edge location -- the default),
    # "PriceClass_200" (excludes South America + Australia/New Zealand),
    # or "PriceClass_100" (North America + Europe only, the cheapest).
    # Viewers outside the selected class are served from the nearest
    # included edge -- functional everywhere, just slower far away.
    # Empty keeps PriceClass_All. The Terraform provider nominally also
    # admits "None" (a shared SDK enum value for the legacy streaming and
    # multi-tenant surfaces), but AWS rejects it on standard
    # distributions -- the three values here mirror AWS's real contract,
    # deliberately stricter than the provider.
    price_class = optional(string, "")

    # The AWS WAF Web ACL protecting the distribution, by ARN.
    # CloudFront-scope ACLs (scope CLOUDFRONT, which must live in
    # us-east-1) are addressed by ARN — never by the bare web ACL ID.
    # Can reference an AwsWafWebAcl resource.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    web_acl_arn = optional(string, "")

    # The content sources the distribution can route to. Each origin's
    # origin_id is your stable handle -- cache behaviors and origin
    # groups target it by that ID.
    origins = list(object({
      # Your stable handle for this origin -- cache behaviors and origin
      # groups target it (e.g. "s3-assets", "api-backend"). Must be
      # unique within the distribution.
      origin_id = string

      # The DNS name CloudFront connects to. For S3 REST origins use the
      # REGIONAL bucket endpoint ("bucket.s3.us-west-2.amazonaws.com" --
      # the bucket_regional_domain_name output of AwsS3Bucket); for
      # load balancers the LB DNS name; for S3 static websites the
      # website endpoint (as a custom_origin -- website endpoints speak
      # plain HTTP only).
      domain_name = string

      # Optional path CloudFront appends to every origin request, e.g.
      # "/production" to serve from a bucket sub-prefix. Must start with
      # "/" and not end with one.
      origin_path = optional(string, "")

      # How many times CloudFront attempts to connect to the origin per
      # request, 1-3. 0 keeps the AWS default (3).
      connection_attempts = optional(number, 0)

      # Seconds CloudFront waits when establishing each connection
      # attempt, 1-10. 0 keeps the AWS default (10).
      connection_timeout_seconds = optional(number, 0)

      # Headers CloudFront adds to every request it sends to this origin.
      # The classic use is a shared-secret header (e.g. "X-Origin-Verify")
      # the origin checks so it only serves CloudFront traffic.
      custom_headers = optional(list(object({
        # The header name, e.g. "X-Origin-Verify".
        name = string

        # The header value. Treat shared-secret values like configuration,
        # not state secrets -- rotate them by updating the distribution.
        value = string
      })), [])

      # Origin Shield: an extra regional caching layer in front of the
      # origin that collapses requests from all edge locations, cutting
      # origin load dramatically for origin-heavy workloads. Billed per
      # request; choose the region closest to the origin.
      origin_shield = optional(object({
        # The AWS region for the Origin Shield cache. Choose the region
        # with the lowest latency to the origin (usually the origin's own
        # region). Example: "us-west-2".
        origin_shield_region = string
      }))

      # S3 REST origin arm: access control for a private bucket.
      s3_origin = optional(object({
        # Create and attach an Origin Access Control for this origin -- the
        # modern way to serve a private bucket (SigV4 signing, works with
        # SSE-KMS and all regions). The distribution's ARN must be allowed
        # in the bucket policy ("AWS:SourceArn" condition on the
        # "cloudfront.amazonaws.com" principal).
        create_origin_access_control = optional(bool, false)

        # Attach an existing legacy Origin Access Identity, as the full
        # "origin-access-identity/cloudfront/<ID>" path. OAI is the
        # predecessor of OAC (no SSE-KMS support, not recommended for new
        # configurations) -- accepted here for buckets already wired to
        # one, never created.
        origin_access_identity = optional(string, "")
      }))

      # Custom (HTTP) origin arm: ports, protocol, and timeouts for load
      # balancers, API endpoints, or any HTTP server.
      custom_origin = optional(object({
        # How CloudFront connects to the origin: "https-only" (recommended),
        # "http-only" (required for S3 website endpoints, which do not speak
        # HTTPS), or "match-viewer" (mirrors the viewer's protocol).
        protocol_policy = string

        # The origin's HTTP port. 0 keeps the default (80).
        http_port = optional(number, 0)

        # The origin's HTTPS port. 0 keeps the default (443).
        https_port = optional(number, 0)

        # TLS versions CloudFront may use to the origin. Empty keeps
        # ["TLSv1.2"] -- the safe modern floor. Only add older protocols for
        # legacy origins that cannot do better.
        ssl_protocols = optional(list(string), [])

        # Seconds an idle connection to the origin is kept open, 1-60 (up
        # to 120 by AWS quota increase). Raising it helps request-heavy
        # workloads reuse connections. 0 keeps the AWS default (5).
        keepalive_timeout_seconds = optional(number, 0)

        # Seconds CloudFront waits for an origin response, 1-60 (up to 180
        # by AWS quota increase). Also the cap on how long a slow API may
        # take before viewers see a 504. 0 keeps the AWS default (30).
        read_timeout_seconds = optional(number, 0)

        # How CloudFront resolves the origin's DNS name: "ipv4" (the AWS
        # default), "ipv6", or "dualstack" (prefer IPv6, fall back to
        # IPv4). Set ipv6/dualstack only for origins that actually publish
        # AAAA records. Empty keeps ipv4.
        ip_address_type = optional(string, "")

        # Mutual TLS toward the ORIGIN: the ACM certificate CloudFront
        # presents as its client certificate when connecting to this origin,
        # by ARN -- the origin verifies it, so only CloudFront (not the open
        # internet) can reach the backend. The certificate must live in
        # us-east-1. Can reference an AwsCertManagerCert resource. Distinct
        # from viewer_mtls, which authenticates VIEWERS to CloudFront.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        mtls_client_certificate_arn = optional(string, "")
      }))

      # VPC origin arm: route to a provisioned CloudFront VPC origin that
      # reaches a private ALB/NLB/instance inside your VPC without any
      # public exposure.
      vpc_origin = optional(object({
        # The VPC origin ID (the CloudFront vpc_origin resource's ID, not a
        # VPC ID).
        vpc_origin_id = string

        # Seconds an idle connection is kept open, 1-120. 0 keeps the AWS
        # default (5).
        keepalive_timeout_seconds = optional(number, 0)

        # Seconds CloudFront waits for a response, 1-180. 0 keeps the AWS
        # default (30).
        read_timeout_seconds = optional(number, 0)

        # The AWS account that owns the VPC origin, for CROSS-ACCOUNT
        # origins (the VPC origin lives in another account that shared it
        # with this one). Empty means the VPC origin belongs to this
        # account.
        owner_account_id = optional(string, "")
      }))

      # Attach an EXISTING Origin Access Control to this origin, by ID.
      # OAC signs origin requests with SigV4 and works with S3, Lambda
      # function URL, MediaPackage v2, and MediaStore origins -- attach a
      # matching-type OAC regardless of which origin arm is set. For S3
      # origins that should get their own OAC created here, use
      # s3_origin.create_origin_access_control instead.
      origin_access_control_id = optional(string, "")

      # Seconds CloudFront keeps the origin connection open waiting for
      # the COMPLETE response (headers plus body). Distinct from the
      # per-read timeout: this caps the whole transfer, catching origins
      # that stream slowly forever. AWS requires it to be >= the origin's
      # read timeout (30 when the read timeout is unset). 0 keeps the AWS
      # default: no completion cap is enforced.
      response_completion_timeout_seconds = optional(number, 0)
    }))

    # Primary/failover origin pairs. A behavior that targets a group's
    # ID gets automatic failover: CloudFront retries the second member
    # when the first returns one of the configured status codes.
    origin_groups = optional(list(object({
      # Your stable handle for this group -- unique across origins and
      # groups.
      origin_group_id = string

      # Exactly two member origin IDs, primary first. Members must be
      # declared origins (groups cannot nest).
      member_origin_ids = list(string)

      # The origin HTTP status codes that trigger failover to the second
      # member, from 400, 403, 404, 416, 500, 502, 503, 504.
      failover_status_codes = list(number)
    })), [])

    # How requests that match no ordered behavior are cached and
    # forwarded -- every distribution has exactly one default behavior.
    default_cache_behavior = object({
      # The origin_id or origin_group_id this behavior routes to.
      target_origin_id = string

      # How viewers reach this content: "redirect-to-https" (the
      # sensible default for websites), "https-only" (APIs; no redirect
      # round-trip), or "allow-all" (legacy plain-HTTP viewers).
      viewer_protocol_policy = string

      # HTTP methods CloudFront forwards to the origin. Empty keeps
      # ["GET", "HEAD"] -- static content. Use the full seven
      # (GET/HEAD/OPTIONS/PUT/POST/PATCH/DELETE) for APIs; only GET/HEAD
      # (+OPTIONS) responses are ever cached.
      allowed_methods = optional(list(string), [])

      # Methods whose responses CloudFront caches -- ["GET", "HEAD"] (the
      # default when empty) or ["GET", "HEAD", "OPTIONS"] (cache CORS
      # preflights).
      cached_methods = optional(list(string), [])

      # Automatically gzip/brotli-compress responses for viewers that
      # accept it -- almost always what you want for text content (the
      # cache policy must enable the compression formats too when using
      # the modern generation).
      compress = optional(bool, false)

      # MODERN generation: the cache policy controlling the cache key and
      # TTLs, by ID -- a managed policy ID (see the message comment) or
      # your own. Mutually exclusive with forwarded_values and the
      # per-behavior TTLs.
      cache_policy_id = optional(string, "")

      # MODERN generation: the origin-request policy controlling which
      # headers/cookies/query strings are forwarded (WITHOUT joining the
      # cache key), by ID.
      origin_request_policy_id = optional(string, "")

      # The response-headers policy adding/removing headers on responses
      # (security headers, CORS, server-timing), by ID. Works with both
      # generations.
      response_headers_policy_id = optional(string, "")

      # LEGACY generation: inline forwarding rules; forwarded values join
      # the cache key. Mutually exclusive with cache_policy_id.
      forwarded_values = optional(object({
        # Forward query strings to the origin (and cache on them).
        query_string = optional(bool, false)

        # When query_string is true, cache only on these parameters (empty
        # caches on all of them).
        query_string_cache_keys = optional(list(string), [])

        # Header names forwarded to the origin and joined into the cache
        # key. "*" forwards all headers and effectively disables caching.
        headers = optional(list(string), [])

        # Cookie forwarding: "none" (the safe default for cacheable
        # content), "whitelist" (only whitelisted_cookie_names), or "all"
        # (destroys cache efficiency -- every cookie combination is its own
        # cache entry).
        cookies_forward = string

        # Cookie names forwarded when cookies_forward is "whitelist".
        whitelisted_cookie_names = optional(list(string), [])
      }))

      # LEGACY generation TTL floor in seconds (with forwarded_values
      # only). Default 0.
      min_ttl_seconds = optional(number, 0)

      # LEGACY generation default TTL in seconds, used when the origin
      # sends no caching headers (with forwarded_values only). 0 keeps
      # the AWS default (86400 -- one day).
      default_ttl_seconds = optional(number, 0)

      # LEGACY generation TTL ceiling in seconds (with forwarded_values
      # only). 0 keeps the AWS default (31536000 -- one year).
      max_ttl_seconds = optional(number, 0)

      # CloudFront Functions (sub-millisecond JavaScript at every edge)
      # attached to this behavior -- viewer-request/viewer-response only,
      # at most one per event type. The lightweight choice for URL
      # rewrites, redirects, and header manipulation.
      function_associations = optional(list(object({
        # The event: "viewer-request" (before the cache lookup) or
        # "viewer-response" (before returning to the viewer).
        event_type = string

        # The CloudFront Function ARN.
        function_arn = string
      })), [])

      # Lambda@Edge functions (full Lambda at regional edges) attached to
      # this behavior -- all four event types, at most one per type. For
      # logic that needs the network, the body, or more than 1ms.
      lambda_function_associations = optional(list(object({
        # The event: "viewer-request"/"viewer-response" (every request; 5s
        # limit) or "origin-request"/"origin-response" (cache misses only;
        # 30s limit -- the usual choice for origin rewrites).
        event_type = string

        # The Lambda function VERSION ARN (Lambda@Edge requires a numbered
        # version, never $LATEST or an alias). The function must live in
        # us-east-1.
        lambda_arn = string

        # Expose the request body to the function (viewer-request and
        # origin-request events).
        include_body = optional(bool, false)
      })), [])

      # Restrict this behavior's content to signed URLs/cookies from
      # these key groups (the modern private-content mechanism).
      trusted_key_group_ids = optional(list(string), [])

      # LEGACY private content: AWS account numbers whose CloudFront key
      # pairs may sign URLs. Prefer trusted_key_group_ids.
      trusted_signers = optional(list(string), [])

      # The field-level encryption configuration ID applied to this
      # behavior (encrypts specific POST fields at the edge with your
      # public key).
      field_level_encryption_id = optional(string, "")

      # The real-time log configuration ARN streaming this behavior's
      # requests to Kinesis within seconds.
      realtime_log_config_arn = optional(string, "")

      # Serve Microsoft Smooth Streaming media from this behavior.
      smooth_streaming = optional(bool, false)

      # Allow gRPC viewer traffic over HTTP/2 (requires POST in
      # allowed_methods and http_version http2 or above).
      grpc_enabled = optional(bool, false)
    })

    # Path-matched behaviors evaluated IN ORDER before the default --
    # first match wins. The classic use: route "/api/*" to a load
    # balancer with caching disabled while everything else serves from
    # S3.
    ordered_cache_behaviors = optional(list(object({
      # The path pattern this behavior matches, e.g. "/api/*",
      # "/images/*.jpg", "*.css". First matching behavior wins.
      path_pattern = string

      # The behavior applied to matching requests.
      behavior = object({
        # The origin_id or origin_group_id this behavior routes to.
        target_origin_id = string

        # How viewers reach this content: "redirect-to-https" (the
        # sensible default for websites), "https-only" (APIs; no redirect
        # round-trip), or "allow-all" (legacy plain-HTTP viewers).
        viewer_protocol_policy = string

        # HTTP methods CloudFront forwards to the origin. Empty keeps
        # ["GET", "HEAD"] -- static content. Use the full seven
        # (GET/HEAD/OPTIONS/PUT/POST/PATCH/DELETE) for APIs; only GET/HEAD
        # (+OPTIONS) responses are ever cached.
        allowed_methods = optional(list(string), [])

        # Methods whose responses CloudFront caches -- ["GET", "HEAD"] (the
        # default when empty) or ["GET", "HEAD", "OPTIONS"] (cache CORS
        # preflights).
        cached_methods = optional(list(string), [])

        # Automatically gzip/brotli-compress responses for viewers that
        # accept it -- almost always what you want for text content (the
        # cache policy must enable the compression formats too when using
        # the modern generation).
        compress = optional(bool, false)

        # MODERN generation: the cache policy controlling the cache key and
        # TTLs, by ID -- a managed policy ID (see the message comment) or
        # your own. Mutually exclusive with forwarded_values and the
        # per-behavior TTLs.
        cache_policy_id = optional(string, "")

        # MODERN generation: the origin-request policy controlling which
        # headers/cookies/query strings are forwarded (WITHOUT joining the
        # cache key), by ID.
        origin_request_policy_id = optional(string, "")

        # The response-headers policy adding/removing headers on responses
        # (security headers, CORS, server-timing), by ID. Works with both
        # generations.
        response_headers_policy_id = optional(string, "")

        # LEGACY generation: inline forwarding rules; forwarded values join
        # the cache key. Mutually exclusive with cache_policy_id.
        forwarded_values = optional(object({
          # Forward query strings to the origin (and cache on them).
          query_string = optional(bool, false)

          # When query_string is true, cache only on these parameters (empty
          # caches on all of them).
          query_string_cache_keys = optional(list(string), [])

          # Header names forwarded to the origin and joined into the cache
          # key. "*" forwards all headers and effectively disables caching.
          headers = optional(list(string), [])

          # Cookie forwarding: "none" (the safe default for cacheable
          # content), "whitelist" (only whitelisted_cookie_names), or "all"
          # (destroys cache efficiency -- every cookie combination is its own
          # cache entry).
          cookies_forward = string

          # Cookie names forwarded when cookies_forward is "whitelist".
          whitelisted_cookie_names = optional(list(string), [])
        }))

        # LEGACY generation TTL floor in seconds (with forwarded_values
        # only). Default 0.
        min_ttl_seconds = optional(number, 0)

        # LEGACY generation default TTL in seconds, used when the origin
        # sends no caching headers (with forwarded_values only). 0 keeps
        # the AWS default (86400 -- one day).
        default_ttl_seconds = optional(number, 0)

        # LEGACY generation TTL ceiling in seconds (with forwarded_values
        # only). 0 keeps the AWS default (31536000 -- one year).
        max_ttl_seconds = optional(number, 0)

        # CloudFront Functions (sub-millisecond JavaScript at every edge)
        # attached to this behavior -- viewer-request/viewer-response only,
        # at most one per event type. The lightweight choice for URL
        # rewrites, redirects, and header manipulation.
        function_associations = optional(list(object({
          # The event: "viewer-request" (before the cache lookup) or
          # "viewer-response" (before returning to the viewer).
          event_type = string

          # The CloudFront Function ARN.
          function_arn = string
        })), [])

        # Lambda@Edge functions (full Lambda at regional edges) attached to
        # this behavior -- all four event types, at most one per type. For
        # logic that needs the network, the body, or more than 1ms.
        lambda_function_associations = optional(list(object({
          # The event: "viewer-request"/"viewer-response" (every request; 5s
          # limit) or "origin-request"/"origin-response" (cache misses only;
          # 30s limit -- the usual choice for origin rewrites).
          event_type = string

          # The Lambda function VERSION ARN (Lambda@Edge requires a numbered
          # version, never $LATEST or an alias). The function must live in
          # us-east-1.
          lambda_arn = string

          # Expose the request body to the function (viewer-request and
          # origin-request events).
          include_body = optional(bool, false)
        })), [])

        # Restrict this behavior's content to signed URLs/cookies from
        # these key groups (the modern private-content mechanism).
        trusted_key_group_ids = optional(list(string), [])

        # LEGACY private content: AWS account numbers whose CloudFront key
        # pairs may sign URLs. Prefer trusted_key_group_ids.
        trusted_signers = optional(list(string), [])

        # The field-level encryption configuration ID applied to this
        # behavior (encrypts specific POST fields at the edge with your
        # public key).
        field_level_encryption_id = optional(string, "")

        # The real-time log configuration ARN streaming this behavior's
        # requests to Kinesis within seconds.
        realtime_log_config_arn = optional(string, "")

        # Serve Microsoft Smooth Streaming media from this behavior.
        smooth_streaming = optional(bool, false)

        # Allow gRPC viewer traffic over HTTP/2 (requires POST in
        # allowed_methods and http_version http2 or above).
        grpc_enabled = optional(bool, false)
      })
    })), [])

    # The certificate presented to viewers. Absent means the default
    # *.cloudfront.net certificate (no aliases possible). Set the ACM
    # arm (recommended) or the IAM arm to serve your own domains --
    # the ACM certificate MUST live in us-east-1.
    viewer_certificate = optional(object({
      # The ACM certificate ARN -- MUST be in us-east-1 regardless of
      # where anything else lives. Can reference an AwsCertManagerCert
      # resource. The recommended arm.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      acm_certificate_arn = optional(string, "")

      # A certificate uploaded to IAM, by ID -- the legacy arm for
      # regions/paths where ACM is unavailable.
      iam_certificate_id = optional(string, "")

      # How CloudFront serves HTTPS for custom certificates: "sni-only"
      # (the default -- free, supported by every modern client),
      # "static-ip" (dedicated static IPs for clients that must pin
      # addresses; contact AWS support to enable, billed), or "vip" (the
      # legacy dedicated-IP tier at significant monthly cost, for ancient
      # non-SNI clients). Empty keeps sni-only.
      ssl_support_method = optional(string, "")

      # The minimum TLS version viewers must speak -- an AWS security
      # policy name. "TLSv1.2_2021" is the module default for custom
      # certificates; "TLSv1.3_2025" (TLS 1.3 only) and "TLSv1.2_2025"
      # are the current-generation policies. "TLSv1.2_2019",
      # "TLSv1.2_2018", "TLSv1.1_2016", "TLSv1_2016", and "TLSv1" remain
      # for older clients, and "SSLv3" exists ONLY for the legacy
      # dedicated-IP tier serving ancient clients -- never choose it for
      # new configurations. Empty keeps TLSv1.2_2021.
      minimum_protocol_version = optional(string, "")
    }))

    # Replace origin error responses with custom pages and control how
    # long errors are cached -- e.g. serve "/errors/404.html" with a
    # 404, or map S3's 403-for-missing-object to a 404.
    custom_error_responses = optional(list(object({
      # The origin HTTP status code to intercept.
      error_code = number

      # The status code returned to the viewer (e.g. map S3's
      # 403-for-missing-object to 404). 0 passes the original code
      # through.
      response_code = optional(number, 0)

      # The page served for this error, as a distribution path (e.g.
      # "/errors/404.html"). The path must be servable by a behavior.
      response_page_path = optional(string, "")

      # Seconds CloudFront caches this error response. 0 keeps the AWS
      # default (300); errors caching too long can mask origin recovery.
      error_caching_min_ttl_seconds = optional(number, 0)
    })), [])

    # Allow or deny viewers by country (ISO 3166-1 alpha-2 codes).
    # Absent means no geographic restriction.
    geo_restriction = optional(object({
      # "whitelist" (serve ONLY the listed countries) or "blacklist"
      # (serve everyone EXCEPT the listed countries).
      restriction_type = string

      # ISO 3166-1 alpha-2 country codes, e.g. "US", "DE", "IN".
      locations = list(string)
    }))

    # Standard access logs delivered to an S3 bucket (best-effort
    # delivery, typically within minutes). Absent disables logging.
    logging = optional(object({
      # The S3 bucket receiving logs, as a bucket DOMAIN NAME
      # ("my-logs.s3.amazonaws.com"). The bucket must have ACLs enabled
      # (object ownership "Bucket owner preferred") -- CloudFront writes
      # logs via the awslogsdelivery canonical user. May be left empty to
      # keep v1 log DELIVERY off while still recording the cookie
      # preference -- the transition shape for distributions moving to
      # v2 (CloudWatch-delivered) access logs, which are configured
      # outside the distribution.
      bucket = optional(string, "")

      # Key prefix for log objects, e.g. "cdn/".
      prefix = optional(string, "")

      # Include cookies in the logs.
      include_cookies = optional(bool, false)
    }))

    # Whether the deployment blocks until CloudFront reports the
    # distribution Deployed at every edge location (typically 5-15
    # minutes). True (the default) means downstream resources see a
    # live distribution; false returns as soon as the configuration is
    # accepted -- the distribution keeps serving the previous
    # configuration until propagation completes.
    wait_for_deployment = optional(bool)

    # Disable the distribution instead of deleting it on destroy. The
    # distribution stops serving but stays in the account for manual
    # deletion later -- an escape hatch for teardown ordering problems;
    # leave false for the normal full delete.
    retain_on_delete = optional(bool, false)

    # Turn on CloudWatch additional metrics (cache hit rate, origin
    # latency, error rates by status code) for the distribution. Billed
    # per the CloudWatch custom-metric rate; indispensable for
    # production cache tuning.
    enable_additional_metrics = optional(bool, false)

    # Mark this distribution as a STAGING distribution -- the target of a
    # blue/green rollout. A staging distribution serves no viewer traffic
    # of its own; a primary distribution's continuous-deployment policy
    # routes a slice of traffic to it by its domain name. Changing this
    # flag REPLACES the distribution (AWS makes it immutable). Staging
    # distributions cannot have aliases -- they are reached only through
    # the primary.
    staging = optional(bool, false)

    # Attach an EXISTING continuous-deployment policy by ID -- for a
    # policy managed outside this resource. To have this distribution
    # own its blue/green policy, use continuous_deployment instead; the
    # two are mutually exclusive.
    continuous_deployment_policy_id = optional(string, "")

    # Create and attach a continuous-deployment policy: route a slice of
    # production traffic to a STAGING distribution (one deployed with
    # staging: true) to validate a configuration change on real traffic
    # before promoting it. The policy is created by this resource and
    # attached to this (primary) distribution.
    continuous_deployment = optional(object({
      # Whether the policy is in effect. True (the default) shifts
      # traffic per the routing choice below; false keeps the policy
      # attached but dormant -- the pause button for a rollout.
      enabled = optional(bool)

      # The CloudFront domain names of the staging distributions
      # receiving the traffic slice (e.g. "d111111abcdef8.cloudfront.net"
      # -- the domain_name output of a distribution deployed with
      # staging: true). CloudFront currently supports one staging
      # distribution per policy.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      staging_distribution_dns_names = list(string)

      # Weight-based routing: send a fixed percentage of ALL traffic to
      # staging. The canary shape. Set exactly one of single_weight or
      # single_header.
      single_weight = optional(object({
        # The fraction of traffic sent to staging, as a decimal between 0
        # and 0.15 (AWS caps staged traffic at 15%). Example: 0.10 sends
        # 10% of requests to the staging distribution.
        weight = number

        # Pin each viewer to one side for the session, so a user does not
        # bounce between primary and staging responses mid-visit. Absent
        # means requests are split without stickiness.
        session_stickiness = optional(object({
          # Seconds of inactivity after which a viewer's session ends,
          # 300-3600. Must not exceed maximum_ttl_seconds.
          idle_ttl_seconds = number

          # The longest a viewer's requests count as one session regardless
          # of activity, 300-3600 seconds.
          maximum_ttl_seconds = number
        }))
      }))

      # Header-based routing: send only requests carrying a specific
      # header/value pair to staging. The team-testing shape -- testers
      # opt in by sending the header; regular viewers never hit staging.
      single_header = optional(object({
        # The request header that opts a request into staging. AWS requires
        # the "aws-cf-cd-" prefix. Example: "aws-cf-cd-canary".
        header = string

        # The header value that must match exactly for the request to route
        # to staging.
        value = string
      }))
    }))

    # The Anycast static IP list this distribution serves from, by ID.
    # Anycast lists give a distribution a small set of dedicated static
    # IPs (for allowlist-style network controls); the feature carries a
    # significant fixed monthly charge and requires the list to be
    # provisioned in the account first.
    anycast_ip_list_id = optional(string, "")

    # The request header CloudFront reads cache tags from (for tag-based
    # invalidation): origin responses carry tags in this header, and an
    # invalidation by tag purges every object labeled with it -- far
    # cheaper than path invalidations for content that clusters by
    # topic. Lowercase only: AWS stores the header name lowercased in the
    # distribution config (live-verified 2026-08-12: "Cache-Tag" read back
    # as "cache-tag") while the provider passes the value through verbatim
    # with no case suppression, so a mixed-case value re-plans as a
    # perpetual cosmetic diff on both engines. HTTP header matching is
    # case-insensitive, so lowercase loses nothing. Example: "cache-tag".
    cache_tag_header_name = optional(string, "")

    # The CloudFront connection function attached to the distribution,
    # by ID. Connection functions run at TCP-connection establishment
    # (before any HTTP parsing) -- the earliest programmable point in
    # the request path.
    connection_function_id = optional(string, "")

    # Mutual TLS for VIEWERS: require or accept client certificates on
    # viewer connections, validated against a CloudFront trust store.
    # The zero-trust front door for machine-to-machine APIs served
    # through CloudFront.
    viewer_mtls = optional(object({
      # The enforcement mode: "required" (reject connections without a
      # valid client certificate), "optional" (request one, admit
      # connections either way -- origins see the validation result in
      # headers), or "passthrough" (forward the certificate to the origin
      # without CloudFront validating it). Empty keeps the AWS default
      # (required).
      mode = optional(string, "")

      # The CloudFront trust store holding the CA bundle that client
      # certificates are validated against, by ID. Required for the
      # "required" and "optional" modes; "passthrough" needs no trust
      # store (the origin does the validating).
      trust_store_id = optional(string, "")

      # Advertise the trust store's CA names in the TLS handshake, so
      # clients holding multiple certificates can pick the right one.
      advertise_trust_store_ca_names = optional(bool, false)

      # Accept client certificates past their expiry date -- a migration
      # escape hatch while a client fleet rotates certificates, never a
      # steady state.
      ignore_certificate_expiry = optional(bool, false)
    }))
  })
}
