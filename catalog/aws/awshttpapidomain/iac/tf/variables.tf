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
  description = "AwsHttpApiDomain specification"
  type = object({
    # The AWS region where the domain name will be created. Must match the
    # region of the certificate and of the APIs mapped onto the domain.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # The fully qualified custom domain name (e.g. "api.example.com").
    # Immutable after creation -- changing it replaces the domain. Must be
    # lowercase. A wildcard domain ("*.example.com") is allowed and matches
    # all first-level subdomains, but requires a certificate that covers the
    # wildcard.
    domain_name = string

    # The ACM certificate for TLS termination on this domain. The certificate
    # must be issued (or imported) in the same region as the domain and must
    # cover domain_name (exact or wildcard match). Accepts a direct certificate
    # ARN or a reference to an AwsCertManagerCert resource.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    certificate_arn = string

    # IP address type for the domain's endpoint.
    # - "ipv4": Resolve to IPv4 addresses only.
    # - "dualstack": Resolve to both IPv4 and IPv6.
    # When omitted, AWS applies its default (dualstack for new domains).
    ip_address_type = optional(string, "")

    # Mutual TLS (mTLS) authentication for the domain. When configured, API
    # Gateway requires clients to present a certificate chaining to a CA in
    # the truststore before any request reaches an API. Common for B2B and
    # machine-to-machine APIs. When mTLS is enabled, also set
    # disable_execute_api_endpoint=true on the mapped APIs -- otherwise
    # callers can bypass mTLS via the default execute-api endpoint.
    mutual_tls = optional(object({
      # S3 URI of the truststore -- a PEM bundle of the CA certificates that
      # client certificates must chain to (e.g.
      # "s3://my-bucket/truststore.pem"). The bucket must be in the same region
      # as the domain.
      truststore_uri = string

      # Optional S3 object version of the truststore. Pin a version so
      # truststore updates are an explicit, auditable change rather than a
      # silent side effect of overwriting the object.
      truststore_version = optional(string, "")
    }))

    # APIs mapped onto this domain. Each mapping binds one API's stage under
    # an optional path key: an empty api_mapping_key serves the API at the
    # domain root ("https://api.example.com/"), while a key like "orders"
    # serves it under "https://api.example.com/orders/". Multiple APIs
    # compose onto one domain by using distinct keys. Mutually exclusive
    # with the routing-rule modes: AWS currently rejects CreateApiMapping
    # for HTTP/WebSocket APIs on a domain whose routing_mode uses routing
    # rules ("APIs with a protocol type of HTTP or WEBSOCKET cannot be
    # associated to domains that have a routingMode that uses
    # RoutingRules", live-verified) -- route through routing_rules instead
    # on such domains.
    api_mappings = optional(list(object({
      # The API to map onto the domain. Accepts a direct API ID or a reference
      # to an AwsHttpApiGateway resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      api_id = string

      # The stage of the API to serve. For HTTP APIs managed by
      # AwsHttpApiGateway this is the stage name exported in the API's outputs
      # (typically "$default").
      stage = string

      # Path key under which the API is served (e.g. "orders" serves the API at
      # "https://<domain>/orders/..."). Leave empty to serve the API at the
      # domain root. Must not contain slashes -- nested keys are not supported
      # by API Gateway v2.
      api_mapping_key = optional(string, "")
    })), [])

    # ARN of an AWS-issued public ACM certificate that proves ownership of the
    # custom domain. Required by AWS in exactly two setups: when
    # certificate_arn is issued by an ACM Private CA, or when mutual_tls is
    # configured with an ACM-IMPORTED certificate. The certificate must be a
    # public ACM certificate for the same domain, in the same region. Accepts
    # a direct ARN or a reference to an AwsCertManagerCert resource.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    ownership_verification_certificate_arn = optional(string, "")

    # How the domain routes incoming requests. Valid values:
    # - "API_MAPPING_ONLY" (AWS default when omitted): only the static
    #   api_mappings path keys route requests.
    # - "ROUTING_RULE_ONLY": only routing_rules route requests.
    # - "ROUTING_RULE_THEN_API_MAPPING": routing rules are evaluated first
    #   (by ascending priority); requests matching no rule fall back to API
    #   mappings. NOTE: AWS currently rejects HTTP/WebSocket API mappings on
    #   rule-mode domains (see api_mappings), so under this spec -- whose
    #   mappings are all HTTP APIs -- the fallback can only serve REST-API
    #   mappings managed outside this resource.
    # Updatable in place.
    routing_mode = optional(string, "")

    # Dynamic routing rules attached to the domain. Each rule matches requests
    # on base path or header values and invokes one REST API stage -- API
    # Gateway supports ONLY REST-protocol targets in routing rules
    # (live-verified; HTTP/WebSocket targets are rejected at
    # CreateRoutingRule). Rules are evaluated in ascending priority order and
    # the first match wins; they route requests only when routing_mode is
    # "ROUTING_RULE_ONLY" or "ROUTING_RULE_THEN_API_MAPPING". Each entry
    # creates one aws_apigatewayv2_routing_rule resource on the domain.
    routing_rules = optional(list(object({
      # Evaluation order of this rule: lower values are evaluated first. Must
      # be between 1 and 1,000,000 and unique across the domain's rules (AWS
      # rejects duplicate priorities). Updatable in place.
      priority = optional(number, 0)

      # Conditions that must ALL match for this rule to fire. Each condition
      # tests exactly one dimension -- a set of candidate base paths, or one
      # header pattern; combine a base-path condition with a header condition
      # to require both.
      conditions = list(object({
        # Candidate base paths (the request's first path segment, without
        # slashes); the condition matches when the request's base path equals ANY
        # entry. Example: ["orders", "billing"]. Case-sensitive.
        base_paths = optional(list(string), [])

        # One header pattern the request must match. The condition matches when
        # the named header's value matches the glob.
        header = optional(object({
          # Header name to test (max 40 characters), e.g. "x-tenant-id".
          # Case-insensitive on the wire, as HTTP headers are.
          name = string

          # Glob pattern the header value must match (max 128 characters), e.g.
          # "tenant-a-*" or an exact literal like "beta".
          value_glob = string
        }))
      }))

      # The REST API that matching requests are routed to. Pass the REST API's
      # ID (AwsRestApiGateway status.outputs.rest_api_id). API Gateway rejects
      # HTTP/WebSocket API targets here, so referencing an AwsHttpApiGateway
      # would always fail the apply.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      api_id = string

      # The stage of the target REST API to invoke (a REST stage name such as
      # "prod").
      stage = string

      # Strip the matched base path from the request before forwarding it to
      # the target API. With "orders" in a rule's base paths and
      # strip_base_path=true, a request to "https://<domain>/orders/list"
      # reaches the API as "/list"; false (the default) forwards
      # "/orders/list" unchanged.
      strip_base_path = optional(bool, false)
    })), [])
  })
}
