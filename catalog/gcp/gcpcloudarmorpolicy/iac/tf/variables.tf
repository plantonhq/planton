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
  description = "GcpCloudArmorPolicy specification"
  type = object({
    # GCP project where the security policy will be created.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the security policy in GCP.
    # Must be 1-63 characters, lowercase letters, numbers, or hyphens.
    # Must start with a lowercase letter and end with a letter or number.
    # If not specified, defaults to metadata.name.
    policy_name = optional(string, "")

    # Description of the security policy. Max 2048 characters.
    description = optional(string, "")

    # Policy type. Determines which features are available and where
    # the policy can be attached.
    #
    # Immutable after creation (ForceNew).
    type = optional(string, "")

    # Adaptive Protection configuration for automatic Layer 7 DDoS detection.
    adaptive_protection_config = optional(object({
      # Enable Cloud Armor Adaptive Protection for Layer 7 DDoS defense.
      # When true, traffic anomalies are detected and alerts are generated.
      enable_layer_7_ddos_defense = optional(bool, false)

      # Rule visibility mode for auto-generated adaptive protection rules.
      # "STANDARD" (default) creates rules visible to all policy viewers.
      # "PREMIUM" requires Cloud Armor Managed Protection Plus.
      rule_visibility = optional(string, "")

      # Per-granularity detection and auto-deploy threshold overrides.
      # Requires enable_layer_7_ddos_defense.
      threshold_configs = optional(list(object({
        # Name of the config. Must be 1-63 characters, RFC1035-compliant, and
        # unique within the policy.
        name = string

        # Confidence threshold (0.0-1.0) an attack signature must reach before
        # auto-deploying a mitigation rule.
        auto_deploy_confidence_threshold = optional(number)

        # Maximum share (0.0-1.0) of baseline (good) traffic the auto-deployed
        # mitigation may impact.
        auto_deploy_impacted_baseline_threshold = optional(number)

        # Load threshold above which auto-deploy considers the backend under
        # attack.
        auto_deploy_load_threshold = optional(number)

        # Lifetime in seconds of an auto-deployed mitigation rule.
        auto_deploy_expiration_sec = optional(number)

        # Detection: absolute queries-per-second considered anomalous.
        detection_absolute_qps = optional(number)

        # Detection: load threshold relative to backend capacity.
        detection_load_threshold = optional(number)

        # Detection: QPS relative to the learned baseline considered anomalous.
        detection_relative_to_baseline_qps = optional(number)

        # Granular traffic units this threshold config applies to.
        traffic_granularity_configs = optional(list(object({
          # Type of granularity: "HTTP_HEADER_HOST" or "HTTP_PATH".
          type = string

          # A specific value of the configured type that constitutes a traffic
          # unit (e.g. one Host name). Leave empty with enable_each_unique_value
          # to treat every unique value as its own unit.
          value = optional(string, "")

          # When true, traffic matching EACH unique value of the type is a
          # separate traffic unit. Only valid when value is empty.
          enable_each_unique_value = optional(bool, false)
        })), [])
      })), [])
    }))

    # Advanced policy-level options: JSON parsing, logging, IP resolution.
    advanced_options_config = optional(object({
      # JSON parsing mode for request body inspection.
      # "DISABLED" (default): No JSON parsing.
      # "STANDARD": Parse JSON bodies for WAF rule evaluation.
      # "STANDARD_WITH_GRAPHQL": Parse JSON and GraphQL bodies.
      json_parsing = optional(string, "")

      # Logging verbosity.
      # "NORMAL" (default): Standard logging.
      # "VERBOSE": Detailed logging including matched rule and request details.
      log_level = optional(string, "")

      # Custom headers to check for the true client IP address.
      # Used when traffic passes through a CDN or reverse proxy that
      # sets the client IP in a custom header. If empty, GCP uses the
      # connection source IP.
      user_ip_request_headers = optional(list(string), [])

      # Additional Content-Type values to parse as JSON. Only meaningful
      # when json_parsing is STANDARD or STANDARD_WITH_GRAPHQL.
      json_custom_config = optional(object({
        # Custom Content-Type header values to parse as JSON
        # (e.g. "application/vnd.api+json").
        content_types = list(string)
      }))

      # How much of each request body the WAF inspects. One of "8KB", "16KB",
      # "32KB", "48KB", "64KB" (provider default 8KB). Larger sizes catch
      # attacks buried deeper in POST bodies at higher processing cost; bodies
      # beyond the limit pass uninspected. Mutable.
      request_body_inspection_size = optional(string, "")
    }))

    # Policy-level reCAPTCHA site key for GOOGLE_RECAPTCHA redirects.
    recaptcha_options_config = optional(object({
      # The reCAPTCHA site key, created from the reCAPTCHA API. The user is
      # responsible for the key's validity.
      redirect_site_key = string
    }))

    # Security rules. Rules are evaluated in priority order (lowest number
    # first). Each rule matches traffic and applies an action.
    #
    # Leave empty to get the API's automatic default "allow all" rule; a
    # non-empty set must include the priority-2147483647 default rule.
    rules = optional(list(object({
      # Action to take when the rule matches.
      # - "allow": Permit the request
      # - "deny(403)": Block with 403 Forbidden
      # - "deny(404)": Block with 404 Not Found
      # - "deny(502)": Block with 502 Bad Gateway
      # - "redirect": Redirect to a configured target (requires redirect_options)
      # - "throttle": Rate-limit the traffic (requires rate_limit_options)
      # - "rate_based_ban": Rate-limit then ban (requires rate_limit_options)
      action = string

      # Rule priority. Lower values are evaluated first.
      # Range: 0 to 2147483647. Each rule must have a unique priority.
      # Priority 2147483647 is the default rule (match "*").
      priority = number

      # Traffic-matching condition. Defines which requests this rule applies to.
      match = object({
        # Predefined match expression. The only supported value is "SRC_IPS_V1",
        # which matches traffic based on source IP address ranges.
        # When set, src_ip_ranges must also be provided.
        versioned_expr = optional(string, "")

        # Source IP CIDR ranges to match against.
        # Required when versioned_expr is "SRC_IPS_V1". Max 10 ranges per rule.
        # Use "*" to match all IP addresses.
        # Examples: ["192.168.1.0/24", "10.0.0.0/8"] or ["*"]
        src_ip_ranges = optional(list(string), [])

        # CEL expression for advanced matching. Supports request attributes
        # such as origin.region_code, request.headers['X-Custom'], request.path,
        # inIpRange(origin.ip, '1.2.3.0/24'), and more.
        # Mutually exclusive with versioned_expr.
        # Example: "origin.region_code == 'US'" or "request.path.matches('/api/.*')"
        expression = optional(string, "")

        # reCAPTCHA site-key options for expressions that evaluate reCAPTCHA
        # tokens. Only meaningful together with expression.
        expr_options = optional(object({
          # Site keys used to validate reCAPTCHA action-tokens.
          action_token_site_keys = optional(list(string), [])

          # Site keys used to validate reCAPTCHA session-tokens.
          session_token_site_keys = optional(list(string), [])
        }))
      })

      # Human-readable description of the rule (max 64 characters).
      description = optional(string, "")

      # If true, the rule is in preview mode: matched traffic is logged but
      # the action is not enforced. Use preview to test rules before enabling.
      preview = optional(bool, false)

      # Rate limit configuration. Required when action is "throttle" or
      # "rate_based_ban".
      rate_limit_options = optional(object({
        # Action to take when traffic is below the threshold. Must be "allow".
        conform_action = string

        # Action to take when traffic exceeds the threshold.
        # Valid values: "redirect", "deny(403)", "deny(404)", "deny(429)", "deny(502)".
        exceed_action = string

        # Single key on which to enforce the rate limit. Determines how requests
        # are grouped for counting. If empty (and no enforce_on_key_configs),
        # defaults to "ALL" (single counter for all matched traffic).
        #
        # Common values:
        #   - "ALL": Single counter for all traffic
        #   - "IP": Per source IP
        #   - "HTTP_HEADER": Per value of a specific header (set enforce_on_key_name)
        #   - "XFF_IP": Per IP from X-Forwarded-For header
        #   - "HTTP_COOKIE": Per cookie value (set enforce_on_key_name)
        #   - "HTTP_PATH": Per URL path
        #   - "SNI": Per TLS Server Name Indication
        #   - "REGION_CODE": Per client country/region
        # Mutually exclusive with enforce_on_key_configs.
        enforce_on_key = optional(string, "")

        # Name of the HTTP header or cookie when enforce_on_key is
        # HTTP_HEADER or HTTP_COOKIE.
        enforce_on_key_name = optional(string, "")

        # Composite rate-limit key: the listed components' values are
        # concatenated to form the key requests are counted against
        # (e.g. IP + HTTP_PATH limits each client per path).
        # Mutually exclusive with enforce_on_key.
        enforce_on_key_configs = optional(list(object({
          # The key type this component contributes.
          enforce_on_key_type = string

          # Name of the HTTP header or cookie when enforce_on_key_type is
          # HTTP_HEADER or HTTP_COOKIE.
          enforce_on_key_name = optional(string, "")
        })), [])

        # Rate limit threshold: when the request count exceeds this value
        # within the interval, the exceed_action is applied.
        rate_limit_threshold = object({
          # Number of requests that triggers the threshold.
          count = number

          # Window of time in seconds over which the count is measured.
          # The API accepts a fixed set of windows.
          interval_sec = number
        })

        # Ban threshold for rate_based_ban actions. When traffic exceeds
        # this threshold after already exceeding the rate_limit_threshold,
        # the source is banned entirely.
        ban_threshold = optional(object({
          # Number of requests that triggers the threshold.
          count = number

          # Window of time in seconds over which the count is measured.
          # The API accepts a fixed set of windows.
          interval_sec = number
        }))

        # Duration of the ban in seconds when using rate_based_ban.
        # Range: 60 to 86400 (1 minute to 24 hours).
        ban_duration_sec = optional(number, 0)

        # Redirect configuration when exceed_action is "redirect".
        exceed_redirect_options = optional(object({
          # Redirect type. EXTERNAL_302 sends a 302 redirect to the target URL.
          # GOOGLE_RECAPTCHA redirects to a Google reCAPTCHA challenge page
          # (customize the site key with the policy-level recaptcha_options_config).
          type = string

          # Target URL for EXTERNAL_302 redirects. Required when type is
          # EXTERNAL_302; must not be set when type is GOOGLE_RECAPTCHA.
          target = optional(string, "")
        }))
      }))

      # Redirect configuration. Required when action is "redirect".
      redirect_options = optional(object({
        # Redirect type. EXTERNAL_302 sends a 302 redirect to the target URL.
        # GOOGLE_RECAPTCHA redirects to a Google reCAPTCHA challenge page
        # (customize the site key with the policy-level recaptcha_options_config).
        type = string

        # Target URL for EXTERNAL_302 redirects. Required when type is
        # EXTERNAL_302; must not be set when type is GOOGLE_RECAPTCHA.
        target = optional(string, "")
      }))

      # Custom headers to inject into matching requests before forwarding
      # to the backend. Only supported for CLOUD_ARMOR type policies.
      header_action = optional(object({
        # Headers to add to matching requests.
        request_headers_to_adds = list(object({
          # HTTP header name.
          header_name = string

          # HTTP header value. If the header already exists, it is overwritten.
          header_value = optional(string, "")
        }))
      }))

      # Preconfigured WAF rule exclusions. Use this to carve out exceptions
      # for specific request fields that trigger false positives in WAF rules.
      # Only supported for CLOUD_ARMOR type policies.
      preconfigured_waf_config = optional(object({
        # Exclusions from preconfigured WAF rules.
        exclusions = list(object({
          # Target WAF rule set to exclude from. Uses the ModSecurity rule set
          # identifiers such as "sqli-v33-stable", "xss-v33-stable",
          # "rce-v33-stable", "lfi-v33-stable", etc.
          target_rule_set = string

          # Specific rule IDs within the rule set to exclude. If empty,
          # the exclusion applies to all rules in the set.
          target_rule_ids = optional(list(string), [])

          # Request headers to exclude from WAF evaluation.
          request_headers = optional(list(object({
            # Comparison operator for matching the field.
            operator = string

            # Value to match against. Required unless operator is EQUALS_ANY
            # (which matches any value for the given field).
            value = optional(string, "")
          })), [])

          # Request cookies to exclude from WAF evaluation.
          request_cookies = optional(list(object({
            # Comparison operator for matching the field.
            operator = string

            # Value to match against. Required unless operator is EQUALS_ANY
            # (which matches any value for the given field).
            value = optional(string, "")
          })), [])

          # Request URIs to exclude from WAF evaluation.
          request_uris = optional(list(object({
            # Comparison operator for matching the field.
            operator = string

            # Value to match against. Required unless operator is EQUALS_ANY
            # (which matches any value for the given field).
            value = optional(string, "")
          })), [])

          # Request query parameters to exclude from WAF evaluation.
          request_query_params = optional(list(object({
            # Comparison operator for matching the field.
            operator = string

            # Value to match against. Required unless operator is EQUALS_ANY
            # (which matches any value for the given field).
            value = optional(string, "")
          })), [])
        }))
      }))
    })), [])

    # User labels attached to the security policy, merged with Planton's
    # platform labels (which win on key conflicts). Mutable.
    labels = optional(map(string), {})

    # Deletion policy for the security policy — what happens on destroy:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the policy and every rule in it are deleted (GCP
    #                refuses while a backend service still attaches it);
    #                the traffic it filtered flows unfiltered once detached
    #   "PREVENT" -- destroy FAILS; protects the WAF/DDoS shield in front
    #                of production backends
    #   "ABANDON" -- the policy is removed from management but keeps
    #                enforcing in GCP
    deletion_policy = optional(string, "")
  })
}
