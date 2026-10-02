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
  description = "GcpMonitoringUptimeCheck specification"
  type = object({
    # The GCP project that owns the uptime check. Can be a literal project ID
    # or a reference to a GcpProject resource. If omitted, the provider's
    # default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Human-friendly name shown in the Cloud Monitoring console. Defaults to
    # metadata.name when left empty (the GCP API requires a display name).
    display_name = optional(string, "")

    # Maximum time to wait for the probe to complete before recording a
    # failure, as a duration in seconds (e.g. "10s"). Must be between 1s and
    # 60s — the GCP API's own bounds.
    timeout = string

    # How often the check runs (default 300s). The GCP API accepts only
    # 60s (1 min), 300s (5 min), 600s (10 min), and 900s (15 min) — the list
    # the provider documents for this field.
    period = optional(string, "")

    # Where the probes originate (default STATIC_IP_CHECKERS — Google's
    # public checker fleet with published static IPs, the right choice for
    # internet-facing targets). VPC_CHECKERS probes private targets from
    # inside your VPC (requires a private checker setup and is chosen
    # automatically by GCP for private targets).
    checker_type = optional(string, "")

    # Regions the check runs from (e.g. USA, EUROPE, SOUTH_AMERICA,
    # ASIA_PACIFIC, or the finer USA_OREGON/USA_IOWA/USA_VIRGINIA). GCP
    # requires enough regions to cover at least 3 checker locations; leaving
    # the list empty runs the check from ALL regions — the recommended
    # default for availability monitoring.
    selected_regions = optional(list(string), [])

    # Whether failed probes are written to Cloud Logging (default false).
    # Enable it when diagnosing flaky checks — each failure logs the probe's
    # observed status and latency.
    log_check_failures = optional(bool, false)

    # A public URL or any monitored resource as the probe target. For the
    # common "is my site up" case, use type uptime_url with labels host
    # (e.g. example.com) and project_id.
    monitored_resource = optional(object({
      # The monitored-resource type (see the message comment for common
      # values). GCP validates the type and its label schema server-side.
      type = string

      # The labels identifying the concrete resource of that type — which
      # labels are required depends on the type (uptime_url needs host;
      # gce_instance needs instance_id and zone; all types accept project_id).
      labels = optional(map(string), {})
    }))

    # A Cloud Monitoring resource GROUP as the probe target — every member
    # of the group is checked. Groups are created in Cloud Monitoring
    # (outside this kind); reference one by its group ID.
    resource_group = optional(object({
      # The group ID (the last segment of the group's resource name
      # projects/{p}/groups/{group_id}). At least one of group_id or
      # resource_type must be set — the provider's own constraint.
      group_id = optional(string, "")

      # What the group's members are: INSTANCE (Compute Engine or AWS EC2) or
      # AWS_ELB_LOAD_BALANCER.
      resource_type = optional(string, "")
    }))

    # A synthetic monitor: GCP invokes the referenced 2nd-gen Cloud Function
    # on the check cadence, and the function's own assertions decide
    # pass/fail. The function carries the probe logic, so http_check and
    # tcp_check are forbidden with this target.
    synthetic_monitor = optional(object({
      # The fully qualified resource name of the 2nd-gen Cloud Function:
      #   projects/{project}/locations/{region}/functions/{name}
      # Can be a literal or a reference to a GcpCloudFunction resource.
      # Immutable: changing the target function replaces the uptime check.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      cloud_function = string
    }))

    # HTTP(S) probe configuration — path, port, TLS, auth, expected status
    # codes.
    http_check = optional(object({
      # URL path to probe (default "/"). A missing leading slash is prepended
      # by GCP.
      path = optional(string, "")

      # Port to probe. Defaults to 80 without SSL and 443 with SSL. Ports 1024
      # and below require use_ssl for STATIC_IP_CHECKERS.
      port = optional(number, 0)

      # HTTP method (default GET). POST requires a content_type and typically a
      # body.
      request_method = optional(string, "")

      # Probe over HTTPS instead of HTTP (default false). Required when the
      # target only serves TLS; also required by GCP for low ports on the
      # public checker fleet.
      use_ssl = optional(bool, false)

      # Verify the target's TLS certificate chain (default false). Only
      # meaningful with use_ssl — GCP ignores it on plain HTTP. Enable it in
      # production so an expired certificate fails the probe instead of passing
      # silently.
      validate_ssl = optional(bool, false)

      # Request body for POST probes, base64-encoded (e.g.
      # base64("foo=bar") = "Zm9vPWJhcg=="). GCP rejects a body on GET probes.
      # Setting a body without a content_type fails the API's own validation.
      body = optional(string, "")

      # How the body is declared in the request's Content-Type header:
      # URL_ENCODED (application/x-www-form-urlencoded) or USER_PROVIDED (the
      # custom_content_type value below).
      content_type = optional(string, "")

      # The Content-Type header value sent when content_type is USER_PROVIDED
      # (e.g. "application/json").
      custom_content_type = optional(string, "")

      # Headers to send with the probe (e.g. a Host header for name-based
      # virtual hosting, or an API key). At most 100 headers.
      headers = optional(map(string), {})

      # Hide header values in the console and API responses (default false).
      # GCP sets this permanently once enabled — turning it back off requires
      # recreating the check. Enable it whenever headers carry credentials.
      mask_headers = optional(bool, false)

      # HTTP basic authentication for the probe.
      auth_info = optional(object({
        # Basic-auth username.
        username = string

        # Basic-auth password — a secret: the platform stores it as a
        # managed-secret reference and resolves it just-in-time at deploy.
        password = optional(string, "")
      }))

      # Authenticate the probe AS the check's Monitoring service agent using an
      # OIDC identity token — the keyless way to probe endpoints (Cloud Run,
      # Cloud Functions) that require authenticated invocations.
      service_agent_authentication = optional(object({
        # The authentication mechanism. OIDC_TOKEN is the only type GCP currently
        # supports.
        type = optional(string, "")
      }))

      # Response status codes that count as SUCCESS. Empty means "2xx only"
      # (GCP's default). Each entry is a class (e.g. STATUS_CLASS_2XX) or one
      # exact status_value — useful when a health endpoint deliberately returns
      # 401/403 to anonymous probes.
      accepted_response_status_codes = optional(list(object({
        # A status class: STATUS_CLASS_1XX, STATUS_CLASS_2XX, STATUS_CLASS_3XX,
        # STATUS_CLASS_4XX, STATUS_CLASS_5XX, or STATUS_CLASS_ANY.
        status_class = optional(string, "")

        # One exact HTTP status code (e.g. 401).
        status_value = optional(number, 0)
      })), [])

      # Include ICMP pings ahead of the HTTP probe (1-3 pings), recording ping
      # latency alongside the check result.
      ping_config = optional(object({
        # Number of ICMP pings to send ahead of the probe (1-3).
        pings_count = optional(number, 0)
      }))
    }))

    # Plain TCP connect probe — succeeds when the port accepts a connection.
    tcp_check = optional(object({
      # Port to connect to. Required — TCP checks have no protocol default.
      port = optional(number, 0)

      # Include ICMP pings ahead of the TCP probe (1-3 pings).
      ping_config = optional(object({
        # Number of ICMP pings to send ahead of the probe (1-3).
        pings_count = optional(number, 0)
      }))
    }))

    # Assertions on the response body. All matchers must pass for the probe
    # to pass. Applies to http_check (response body) and tcp_check (bytes
    # read); leave empty to assert only on status/connectivity.
    content_matchers = optional(list(object({
      # The string, regex, or JSON-path expectation to test the response
      # against.
      content = string

      # How `content` is interpreted (default CONTAINS_STRING).
      matcher = optional(string, "")

      # JSON-path details for the MATCHES_JSON_PATH / NOT_MATCHES_JSON_PATH
      # matchers.
      json_path_matcher = optional(object({
        # The JSONPath expression selecting the value to test (e.g.
        # "$.status" or "$.items[0].state").
        json_path = string

        # How the selected value is compared to `content`: EXACT_MATCH (default)
        # or REGEX_MATCH.
        json_matcher = optional(string, "")
      }))
    })), [])

    # User labels attached to the uptime check for organizing and identifying
    # it (maps to the provider's user_labels), merged with Planton's platform
    # labels (which win on key conflicts).
    labels = optional(map(string), {})

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the uptime check is deleted; alert policies that filter
    #                on its uptime_check_id stop receiving data
    #   "PREVENT" -- destroy FAILS; protects production availability
    #                monitoring from accidental teardown
    #   "ABANDON" -- the check is removed from management but keeps running
    #                (and billing) in GCP
    deletion_policy = optional(string, "")
  })
}
