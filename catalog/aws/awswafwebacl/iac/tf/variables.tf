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
  description = "AwsWafWebAcl specification"
  type = object({
    # The AWS region where the resource will be created.
    # For CLOUDFRONT scope this must be "us-east-1" (the WAF global region).
    # Example: "us-west-2", "us-east-1"
    region = string

    # Scope determines where the Web ACL can be used. Create-time immutable
    # (ForceNew) — changing it replaces the web ACL.
    #
    # "REGIONAL" — protects ALB, API Gateway, AppSync, Cognito, App Runner,
    # and Verified Access resources in the web ACL's own region.
    #
    # "CLOUDFRONT" — protects CloudFront distributions. The Web ACL MUST be
    # created in us-east-1 (region must be "us-east-1").
    scope = string

    # Action to take when no rule matches a request. This is the "baseline"
    # security posture:
    # - "allow" (permissive): allow all traffic unless a rule blocks it.
    #   Use when most traffic is legitimate (e.g., public website).
    # - "block" (restrictive): block all traffic unless a rule allows it.
    #   Use when most traffic should be denied (e.g., private API).
    default_action = object({
      # Action type: "allow" or "block".
      type = string

      # Custom response configuration for block actions. Only valid when
      # type is "block". Specifies the HTTP response code and optional body
      # to return to blocked requests.
      custom_response = optional(object({
        # HTTP response status code to return. Range: 200-600.
        # Common values: 403 (Forbidden), 429 (Too Many Requests), 503 (Service Unavailable).
        response_code = number

        # Key referencing a custom_response_body defined at the Web ACL level.
        # When set, the response body from the matching custom_response_body is
        # returned with the specified response_code and content type.
        custom_response_body_key = optional(string, "")

        # Additional HTTP headers to include in the block response.
        response_headers = optional(list(object({
          # HTTP header name (case-insensitive). 1-64 characters: letters, digits,
          # and _ $ . - only. For INSERTED request headers, WAF prefixes the name
          # with "x-amzn-waf-" on the wire (name "sample" arrives as
          # "x-amzn-waf-sample") to avoid clobbering existing request headers;
          # response headers are sent under the name as given.
          name = string

          # HTTP header value. 1-255 characters.
          value = string
        })), [])
      }))

      # Custom request headers to insert for allow actions. Only valid when
      # type is "allow". Headers are added to the request before forwarding
      # to the protected resource.
      custom_request_headers = optional(list(object({
        # HTTP header name (case-insensitive). 1-64 characters: letters, digits,
        # and _ $ . - only. For INSERTED request headers, WAF prefixes the name
        # with "x-amzn-waf-" on the wire (name "sample" arrives as
        # "x-amzn-waf-sample") to avoid clobbering existing request headers;
        # response headers are sent under the name as given.
        name = string

        # HTTP header value. 1-255 characters.
        value = string
      })), [])
    })

    # Human-readable description of the Web ACL. AWS restricts the character
    # set: letters, digits, whitespace, and _ + = : # @ / - , . only (notably
    # NO parentheses), 3-256 characters — WAF rejects anything else at create
    # time, so the constraint is enforced here where the failure is immediate
    # and readable.
    description = optional(string, "")

    # Ordered set of rules evaluated against each incoming request. Rules are
    # evaluated by priority (lowest number first). When a rule matches, its
    # action is taken and evaluation stops for that request.
    #
    # When no rules are provided, only the default_action applies. This is
    # valid but uncommon — most Web ACLs have at least one managed rule group.
    rules = optional(any, [])

    # CloudWatch metrics configuration for the Web ACL itself.
    #
    # When omitted, the IaC module applies sensible defaults:
    # - cloudwatch_metrics_enabled = true
    # - sampled_requests_enabled = true
    # - metric_name = resource name
    visibility_config = optional(object({
      # Enable CloudWatch metrics for this Web ACL or rule.
      # Default: true (applied by IaC module when omitted).
      cloudwatch_metrics_enabled = optional(bool, false)

      # Enable request sampling for this Web ACL or rule. Sampled requests are
      # viewable in the AWS WAF console for debugging rule matches.
      # Default: true (applied by IaC module when omitted).
      sampled_requests_enabled = optional(bool, false)

      # CloudWatch metric name for this Web ACL or rule. Must be unique within
      # the Web ACL's rules. 1-128 characters: letters, digits, hyphen,
      # underscore only; "All" and "Default_Action" are reserved by WAF.
      # Default: resource name (Web ACL) or rule name (applied by IaC module).
      metric_name = optional(string, "")
    }))

    # Reusable response body templates that can be referenced by block actions
    # via custom_response_body_key. Define branded error pages or structured
    # error responses here and reference them by key from individual rules.
    #
    # Each entry requires a unique key (used as the reference), content string,
    # and content_type.
    custom_response_bodies = optional(list(object({
      # Unique key used to reference this response body from custom_response
      # configurations. Must be unique within the Web ACL. 1-128 characters:
      # letters, digits, underscore, hyphen.
      key = string

      # The response body content (HTML, plain text, or JSON). Max 10,240 bytes.
      content = string

      # MIME type of the content.
      # Valid values: "TEXT_PLAIN", "TEXT_HTML", "APPLICATION_JSON".
      content_type = string
    })), [])

    # Domains to accept in web request tokens for CAPTCHA and Challenge actions.
    # Required when using CAPTCHA/Challenge with multiple domains that share
    # the same Web ACL. When omitted, tokens are scoped to the request domain.
    # Each entry is a domain name (1-253 characters; letters, digits, dots,
    # hyphens, slashes). Public suffixes (e.g. "co.uk", "gov.au") are rejected
    # by AWS.
    token_domains = optional(list(string), [])

    # How long a client's successful CAPTCHA solve remains valid, web-ACL-wide
    # (immunity time in seconds, 60–259,200; AWS default 300). Rules can
    # override this per rule via the rule's captcha_config.
    captcha_config = optional(object({
      # Immunity time in seconds. CAPTCHA: 60–259,200 (AWS default 300).
      # Challenge: 300–259,200 (AWS default 300). Typed int32 (the max fits) —
      # protojson stringifies 64-bit integers, which would corrupt the rule-JSON
      # document for the per-rule configs living inside the rules subtree.
      immunity_time_sec = number
    }))

    # How long a client's successful silent-challenge response remains valid,
    # web-ACL-wide (immunity time in seconds, 300–259,200; AWS default 300).
    # Rules can override this per rule via the rule's challenge_config.
    challenge_config = optional(object({
      # Immunity time in seconds. CAPTCHA: 60–259,200 (AWS default 300).
      # Challenge: 300–259,200 (AWS default 300). Typed int32 (the max fits) —
      # protojson stringifies 64-bit integers, which would corrupt the rule-JSON
      # document for the per-rule configs living inside the rules subtree.
      immunity_time_sec = number
    }))

    # Per-resource-type request BODY inspection size limits. By default WAF
    # inspects only the first 16 KB of a request body; raise the limit (to 32,
    # 48, or 64 KB) for APIs that carry large JSON payloads whose tail must
    # still be inspected. Larger limits increase WCU cost of body-inspecting
    # rules. CloudFront also supports raising this; the other resource types
    # are capped per their entry.
    association_config = optional(object({
      # Body inspection limit for CloudFront distributions ("KB_16", "KB_32",
      # "KB_48", or "KB_64").
      cloudfront_request_body_limit = optional(string, "")

      # Body inspection limit for API Gateway REST APIs.
      api_gateway_request_body_limit = optional(string, "")

      # Body inspection limit for Cognito user pools.
      cognito_user_pool_request_body_limit = optional(string, "")

      # Body inspection limit for App Runner services.
      app_runner_service_request_body_limit = optional(string, "")

      # Body inspection limit for Verified Access instances.
      verified_access_instance_request_body_limit = optional(string, "")
    }))

    # Field-level data protection: replace or hash specified request fields
    # (headers, cookies, query strings, body) in ALL WAF outputs — logs,
    # sampled requests, and rule match details — before they leave WAF. This
    # is stronger than the logging block's redacted fields (which only affect
    # the logging destination): use it for PII that must never appear anywhere.
    data_protection_config = optional(object({
      # The fields to protect (up to 26 entries).
      data_protections = list(object({
        # The field class to protect: "SINGLE_HEADER", "SINGLE_COOKIE",
        # "SINGLE_QUERY_ARGUMENT", "QUERY_STRING", or "BODY".
        field_type = string

        # The specific keys to protect for the SINGLE_* field types (header names,
        # cookie names, query-argument names; up to 100, each 1-64 characters).
        # OMITTING keys for a SINGLE_* type protects ALL keys of that type (the
        # AWS contract: "If you don't specify any key, then all keys for the
        # field type are protected"). Not applicable to QUERY_STRING/BODY (which
        # protect the whole component).
        field_keys = optional(list(string), [])

        # How to mask: "SUBSTITUTION" (replace with a fixed placeholder) or
        # "HASH" (replace with a one-way hash, preserving correlate-ability).
        action = string

        # Also exclude this field from rule MATCH details in logs.
        exclude_rule_match_details = optional(bool, false)

        # Also exclude this field from rate-based rule details in logs.
        exclude_rate_based_details = optional(bool, false)
      }))
    }))

    # Optional logging configuration. When provided, WAF sends detailed request
    # logs to the specified destination (CloudWatch Logs, S3, or Kinesis
    # Firehose).
    #
    # Important naming constraint: the destination resource name must start with
    # "aws-waf-logs-" (enforced by AWS). For example:
    # - CloudWatch Log Group: "aws-waf-logs-my-acl"
    # - S3 Bucket: "aws-waf-logs-my-acl"
    # - Kinesis Firehose: "aws-waf-logs-my-acl"
    logging = optional(object({
      # ARN of the logging destination. Must be one of:
      # - CloudWatch Logs log group ARN
      # - S3 bucket ARN
      # - Kinesis Firehose delivery stream ARN
      #
      # The destination resource name must start with "aws-waf-logs-".
      #
      # Deliberately singular: AWS allows exactly one logging destination per
      # web ACL (the Terraform provider's log_destination_configs argument
      # nominally accepts up to 100 entries, but the service contract — SDK:
      # "You can associate one logging destination to a web ACL" — is one).
      #
      # No default_kind is set because the destination can be any of three
      # different resource types.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      destination_arn = string

      # HTTP header names to redact from logs (each 1-64 characters). Redacted
      # headers appear as "REDACTED" in log entries instead of their actual
      # values.
      #
      # Common headers to redact: "Authorization", "Cookie", "X-Api-Key".
      redacted_header_names = optional(list(string), [])

      # Redact the URI path from log entries. When true, the URI path appears
      # as "REDACTED" in logs.
      redact_uri_path = optional(bool, false)

      # Redact the query string from log entries. When true, query string
      # parameters appear as "REDACTED" in logs.
      redact_query_string = optional(bool, false)

      # Redact the HTTP method from log entries. When true, the method appears
      # as "REDACTED" in logs. (Method, query string, URI path, and single
      # headers are the only components AWS supports redacting.)
      redact_method = optional(bool, false)

      # Optional log filtering: keep or drop log records by the action WAF
      # applied or by labels on the request, instead of logging every inspected
      # request. Typical use: keep only BLOCK and COUNT records to cut logging
      # cost on high-traffic ACLs.
      filter = optional(object({
        # What happens to log records that match NONE of the filters:
        # "KEEP" (log them) or "DROP" (discard them).
        default_behavior = string

        # The filters, each with its own keep/drop behavior. At least one.
        filters = list(object({
          # What happens to log records matching this filter: "KEEP" or "DROP".
          behavior = string

          # How the conditions combine: "MEETS_ALL" (every condition must match)
          # or "MEETS_ANY" (at least one condition matches).
          requirement = string

          # The match conditions. At least one.
          conditions = list(object({
            # Match log records by the action WAF applied to the request:
            # "ALLOW", "BLOCK", "COUNT", "CAPTCHA", "CHALLENGE",
            # "EXCLUDED_AS_COUNT" (rules overridden to count — the tuning-noise
            # filter), or "MONETIZE" (marketplace rule groups only).
            action = optional(string, "")

            # Match log records carrying this label. Must be the FULLY QUALIFIED
            # label name including the prefix, e.g.
            # "awswaf:managed:aws:bot-control:bot:category:monitoring".
            label_name = optional(string, "")
          }))
        }))
      }))
    }))
  })
}
