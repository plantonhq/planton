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
  description = "AwsRoute53HealthCheck specification"
  type = object({
    # The AWS region where the resource will be created.
    # Route 53 health checks are global objects; this selects the region used
    # for provider API calls.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # The monitoring model (create-time immutable, ForceNew):
    # - "HTTP" / "HTTPS": healthy when the endpoint answers with 2xx/3xx
    #   within the timeout.
    # - "HTTP_STR_MATCH" / "HTTPS_STR_MATCH": additionally requires
    #   search_string to appear in the first 5,120 bytes of the response body.
    # - "TCP": healthy when a TCP connection can be established.
    # - "CALCULATED": aggregates child_health_checks.
    # - "CLOUDWATCH_METRIC": mirrors a CloudWatch alarm's state.
    # - "RECOVERY_CONTROL": mirrors an Application Recovery Controller routing
    #   control state.
    check_type = string

    # Domain name of the endpoint to probe. For HTTP(S) checks without
    # ip_address, Route 53 resolves this name and probes the result (also sent
    # as the Host header). When ip_address is also set, the probe goes to the
    # IP and this value is only the Host header. Max 255 characters.
    fqdn = optional(string, "")

    # IPv4 or IPv6 address of the endpoint to probe. Use for endpoints whose
    # address is static; use fqdn alone when the address changes (e.g. behind
    # DNS-based scaling). The address must be publicly routable: AWS rejects
    # local, private, non-routable, multicast, AND documentation/reserved
    # ranges at CreateHealthCheck with InvalidInput ("IPv4 address x.x.x.x is
    # forbidden" — proven live against RFC 5737 192.0.2.x, even on a disabled
    # check). The contract is server-side only — no schema mirrors it — so a
    # placeholder manifest must use fqdn instead of a made-up address.
    ip_address = optional(string, "")

    # TCP port of the endpoint. Defaults: 80 for HTTP/HTTP_STR_MATCH, 443 for
    # HTTPS/HTTPS_STR_MATCH. Required for TCP checks (there is no default).
    port = optional(number, 0)

    # Path to probe for HTTP(S) checks (e.g. "/healthz"). Defaults to "/".
    # Not applicable to TCP.
    resource_path = optional(string, "")

    # String that must appear in the first 5,120 bytes of the response body.
    # Required for (and only valid with) HTTP_STR_MATCH / HTTPS_STR_MATCH.
    search_string = optional(string, "")

    # Seconds between probes from each checker: 10 or 30. AWS defaults to 30
    # when omitted — no default is materialized here, because the field only
    # exists for endpoint checks and a manufactured value would be dead
    # configuration on every other type. Create-time immutable (ForceNew).
    # Fast (10s) checks cost more but detect failures ~3x sooner.
    request_interval = optional(number)

    # Consecutive probe results required to flip the health state (1–10; AWS
    # defaults to 3 when omitted). Lower reacts faster; higher rides out
    # blips. Endpoint checks only.
    failure_threshold = optional(number)

    # Measure and graph endpoint latency in the Route 53 console.
    # Create-time immutable (ForceNew); small extra cost.
    measure_latency = optional(bool, false)

    # Send SNI (the fqdn value) in the TLS handshake for HTTPS checks —
    # required by most name-based virtual hosting endpoints. AWS defaults this
    # to true for HTTPS checks when fqdn is set.
    enable_sni = optional(bool)

    # Subset of Route 53 checker regions to probe from (minimum 3 when set).
    # Valid values: us-east-1, us-west-1, us-west-2, eu-west-1, ap-southeast-1,
    # ap-southeast-2, ap-northeast-1, sa-east-1. Default: all checker regions.
    # Note: once set, AWS ignores removing the list (it keeps the last value).
    regions = optional(list(string), [])

    # Invert the result: report unhealthy when the underlying check is healthy
    # and vice versa. Occasionally useful for "route AWAY while X is up"
    # arrangements.
    invert_healthcheck = optional(bool, false)

    # Administratively disable probing. Route 53 then treats the check as
    # always healthy (unless inverted) — the maintenance-window switch that
    # stops failover from firing while you work on the endpoint.
    disabled = optional(bool, false)

    # The child health checks this calculated check aggregates (max 256).
    # Can reference other AwsRoute53HealthCheck resources.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    child_health_checks = optional(list(string), [])

    # Minimum number of healthy children for this check to report healthy
    # (0–256). AWS's contract (per the CreateHealthCheck API): an explicit 0
    # makes the check ALWAYS HEALTHY; a value greater than the number of
    # children makes it always unhealthy; when omitted, Route 53 applies its
    # own server-side default. Explicit 0 and omitted are therefore different
    # configurations — presence carries the distinction.
    child_health_threshold = optional(number)

    # Name of the CloudWatch alarm whose state this check mirrors.
    cloudwatch_alarm_name = optional(string, "")

    # Region the CloudWatch alarm lives in (alarms are regional even though
    # the health check is global). Example: "us-west-2".
    # Format-checked rather than enumerated on purpose: the provider's region
    # enum grows with every AWS region launch; AWS rejects unknown regions.
    cloudwatch_alarm_region = optional(string, "")

    # What to report while the alarm is in INSUFFICIENT_DATA state:
    # "Healthy", "Unhealthy", or "LastKnownStatus".
    insufficient_data_health_status = optional(string, "")

    # ARN of the Application Recovery Controller routing control whose state
    # this check mirrors. Create-time immutable (ForceNew).
    routing_control_arn = optional(string, "")
  })
}
