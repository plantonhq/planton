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
  description = "AwsGlobalAccelerator specification"
  type = object({
    # The AWS provider region used for deployment. Global Accelerator is a
    # GLOBAL service — its control-plane API is homed in us-west-2, and the
    # provider transparently pins API calls there regardless of this value.
    # This region still matters in one place: it is the default
    # endpoint_group_region for any endpoint group that does not set one, so
    # point it at the region where your primary endpoints live.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Whether the accelerator is enabled and accepting traffic. When disabled,
    # the accelerator's DNS name stops resolving and no traffic is routed.
    # Useful for temporarily disabling an accelerator during maintenance without
    # destroying it (the static IPs are retained while disabled).
    enabled = optional(bool)

    # IP address type for the accelerator.
    # - "IPV4": Two static IPv4 anycast addresses (default).
    # - "DUAL_STACK": IPv4 + IPv6 anycast addresses for clients on IPv6 networks.
    #   Dual-stack accelerators additionally export a dual-stack DNS name.
    ip_address_type = optional(string)

    # Bring-Your-Own-IP (BYOIP) addresses to assign to the accelerator instead
    # of AWS-allocated anycast IPs. Provide exactly 1 or 2 IPv4 addresses from
    # a BYOIP address pool registered with AWS (maximum 2 — AWS hard limit).
    #
    # ForceNew — changing this destroys and recreates the accelerator.
    # Leave empty to use AWS-allocated IPs (the default for most deployments).
    ip_addresses = optional(list(string), [])

    # Optional flow log configuration for traffic analysis. When enabled,
    # Global Accelerator publishes flow logs to the specified S3 bucket.
    #
    # Provider quirk (handled by the modules): flow-log settings ride a separate
    # accelerator-attributes API call after create, and changing the bucket or
    # prefix while flow logs are enabled requires AWS to briefly disable and
    # re-enable them — expect two deployment waits for that class of update.
    flow_logs = optional(object({
      # Enable flow log delivery. Setting this to false on an accelerator that
      # previously had flow logs enabled turns them off (the modules always send
      # the explicit disabled state — silence would leave AWS logging forever).
      enabled = optional(bool, false)

      # S3 bucket name for flow log storage. Required when enabled is true.
      # The bucket must exist and grant Global Accelerator write permission.
      #
      # The presence coupling below checks only that the reference is supplied;
      # whether it carries a literal value or resolves through valueFrom is
      # validated by the reference resolver (message-level CEL cannot dereference
      # StringValueOrRef sub-fields — a protovalidate-java constraint).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      s3_bucket = optional(string, "")

      # S3 key prefix for flow logs. Useful for organizing logs when multiple
      # accelerators share a bucket. Example: "ga-logs/prod-accelerator/".
      # Maximum 255 characters (the provider's bound).
      s3_prefix = optional(string, "")
    }))

    # Listeners define the ports and protocols the accelerator accepts traffic on.
    # Each listener routes traffic to one or more regional endpoint groups.
    #
    # At least one listener is required — an accelerator without listeners
    # serves no purpose beyond reserving static IPs.
    listeners = list(object({
      # User-assigned name for this listener. Used as a key in the output maps
      # (listener_arns, endpoint_group_arns) so downstream resources can reference
      # specific listener or endpoint group ARNs via valueFrom. This name is a
      # Planton-side key — AWS listeners have no name of their own.
      #
      # Must be unique within the accelerator's listeners. Lowercase alphanumeric
      # and hyphens only, starting with a letter (max 63 characters).
      name = string

      # Protocol for this listener. Global Accelerator operates at Layer 4.
      # - "TCP": For HTTP, HTTPS, WebSocket, gRPC, and other TCP workloads.
      # - "UDP": For DNS, gaming, IoT, and real-time media workloads.
      protocol = string

      # Client affinity setting for this listener.
      # - "NONE" (default): Requests from the same client may be routed to
      #   different endpoints. Best for stateless workloads.
      # - "SOURCE_IP": All requests from the same source IP address are routed
      #   to the same endpoint within an endpoint group. Required for stateful
      #   protocols (gaming, WebSocket connections, long-lived TCP sessions).
      client_affinity = optional(string)

      # Port ranges that this listener accepts traffic on. Each range defines
      # a from_port and to_port (inclusive). Use a single port range for most
      # workloads, or multiple ranges for services on different ports.
      #
      # At least one port range is required. Maximum 10 ranges per listener
      # (AWS hard limit).
      port_ranges = list(object({
        # First port in the range (inclusive). Range: 1-65535.
        from_port = number

        # Last port in the range (inclusive). Must be >= from_port. For a single
        # port, set from_port and to_port to the same value. Range: 1-65535.
        to_port = number
      }))

      # Endpoint groups define regional destinations for this listener's traffic.
      # Each endpoint group represents a set of endpoints in one AWS region.
      #
      # At least one endpoint group is required — a listener without endpoint
      # groups drops all traffic.
      endpoint_groups = list(object({
        # User-assigned name for this endpoint group. Used as part of the composite
        # key in the endpoint_group_arns output map (format: "listener_name/group_name").
        # This name is a Planton-side key — AWS endpoint groups have no name of
        # their own.
        #
        # Must be unique within the parent listener's endpoint groups. Lowercase
        # alphanumeric and hyphens only, starting with a letter (max 63 characters).
        name = string

        # AWS region for this endpoint group (e.g., "us-east-1", "eu-west-1").
        # When omitted, defaults to the spec's region. ForceNew — changing
        # the region requires replacing the endpoint group. The format is checked
        # here; region-name membership is validated by AWS (the region list grows).
        endpoint_group_region = optional(string, "")

        # Port to use for health checks. When omitted, AWS uses the first port of
        # the listener's port ranges. Set this to check health on a dedicated
        # health-check port separate from the traffic port. Range: 1-65535.
        health_check_port = optional(number)

        # Protocol for health checks.
        # - "TCP" (default): Verifies port reachability only.
        # - "HTTP": Sends GET request to health_check_path, expects 200 response.
        # - "HTTPS": Same as HTTP but over TLS.
        health_check_protocol = optional(string)

        # Path for HTTP/HTTPS health checks. The accelerator sends a GET request
        # to this path (AWS defaults to "/" when omitted). Required here when
        # health_check_protocol is HTTP or HTTPS so intent is explicit. Ignored
        # for TCP health checks. Max 255 characters. Example: "/health".
        health_check_path = optional(string, "")

        # Seconds between health checks for each endpoint. AWS accepts exactly
        # two values: 10 or 30. Default: 30. (The Terraform provider's schema
        # validates the looser 10-30 range, but the Global Accelerator API
        # rejects anything except 10 or 30 at create time — this rule carries
        # the service's real contract so misconfigurations fail at validation,
        # not at deploy.)
        health_check_interval_seconds = optional(number)

        # Number of consecutive health checks that must succeed (or fail) to
        # change an endpoint's health status. Range: 1-10. Default: 3.
        threshold_count = optional(number)

        # Percentage of traffic to route to this endpoint group. Range: 0.0-100.0.
        # When omitted, AWS routes all traffic (100.0). Use values below 100 for
        # gradual traffic shifting between regions (blue/green, canary deployments).
        #
        # Set to 0 explicitly to temporarily drain a region without removing its
        # endpoints — 0 is a real value here, distinct from omitting the field.
        traffic_dial_percentage = optional(number)

        # Endpoints within this regional group. Each endpoint is a resource that
        # receives traffic — an ALB, NLB, Elastic IP, or EC2 instance.
        #
        # Endpoints are optional. You may create the endpoint group first and
        # register endpoints later (e.g., when the ALB or NLB is deployed).
        endpoints = optional(list(object({
          # Resource identifier for the endpoint. Accepts:
          # - ALB ARN: "arn:aws:elasticloadbalancing:..."
          # - NLB ARN: "arn:aws:elasticloadbalancing:..."
          # - EIP allocation ID: "eipalloc-..."
          # - EC2 instance ID: "i-..."
          #
          # Uses StringValueOrRef for cross-resource referencing (e.g., reference an
          # AwsAlb's load_balancer_arn output or an AwsElasticIp's allocation_id
          # output via valueFrom). No default_kind is set because the target resource
          # type varies.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          endpoint_id = string

          # Relative weight for this endpoint. Range: 0-255. When omitted, both
          # modules materialize AWS's documented default of 128 — the provider has no
          # default of its own and would otherwise transmit 0, silently draining the
          # endpoint. Higher weight means more traffic.
          #
          # Set to 0 explicitly to temporarily stop routing traffic to this endpoint
          # without removing it — 0 is a real value here, distinct from omitting the
          # field.
          weight = optional(number)

          # Preserve the client's source IP address in requests forwarded to the
          # endpoint. Applies to Application Load Balancer and EC2 instance
          # endpoints; when omitted, AWS applies its per-endpoint-type default
          # (false for ALB endpoints).
          #
          # Operational note: when enabled, Global Accelerator creates a security
          # group named "GlobalAccelerator" in the endpoint's VPC that must be
          # deleted before that VPC can be destroyed.
          client_ip_preservation_enabled = optional(bool)

          # ARN of a Global Accelerator cross-account attachment that authorizes
          # this endpoint when it lives in another AWS account. Create the
          # attachment in the endpoint-owning account (with this accelerator's
          # account as a principal) and supply its ARN here. Leave empty for
          # same-account endpoints — the common case.
          attachment_arn = optional(string, "")
        })), [])

        # Port overrides remap listener ports to different endpoint ports. Useful
        # when the listener accepts traffic on one port but the endpoint serves
        # on a different port. Maximum 10 overrides per endpoint group (AWS hard
        # limit).
        #
        # Example: listener on port 443, endpoint on port 8443.
        port_overrides = optional(list(object({
          # The listener port that is remapped. Must match a port within one of the
          # listener's port ranges.
          listener_port = number

          # The endpoint port that traffic is forwarded to.
          endpoint_port = number
        })), [])
      }))
    }))
  })
}
