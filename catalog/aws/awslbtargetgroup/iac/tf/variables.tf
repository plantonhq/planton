variable "metadata" {
  description = "Catalog object metadata"
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
  description = "AwsLbTargetGroup specification"
  type = object({
    # The AWS region where the target group is created.
    # Must match the region of the VPC and of any load balancer that forwards
    # to this group. Example: "us-west-2", "eu-west-1".
    region = string

    # The explicit AWS-side target group name. Empty (the common case) means
    # the group is named after `metadata.name`, truncated to 32 characters. Set
    # it when the AWS name must differ from `metadata.name` -- for example a
    # target group taken over from AWS under a generated name, which keeps that
    # name while the component carries a readable one. AWS allows up to 32
    # alphanumeric characters and hyphens, not beginning or ending with a
    # hyphen. ForceNew: a name different from the deployed one replaces the
    # target group.
    target_group_name = optional(string, "")

    # The VPC the targets live in. Required for "instance", "ip", and "alb"
    # target types; ignored for "lambda" (a Lambda function is not addressed
    # through a VPC). Immutable: changing the VPC replaces the target group.
    #
    # Requiredness is enforced by the IaC modules rather than a proto rule:
    # message-level CEL cannot inspect StringValueOrRef fields without breaking
    # protovalidate-java, so both engines fail fast with a clear error when the
    # VPC is missing for a non-lambda target type.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    vpc_id = optional(string, "")

    # How targets are registered into this group. Immutable.
    # - "instance" (default): targets are EC2 instance IDs; NLB health checks
    #   use the instance's primary private IP.
    # - "ip": targets are IP addresses from the VPC CIDR or peered/on-premises
    #   ranges routable through the VPC. The type ECS awsvpc tasks and most
    #   Kubernetes pod-IP integrations use.
    # - "lambda": the single target is a Lambda function; the load balancer
    #   invokes it directly, so port/protocol/vpc_id do not apply.
    # - "alb": the single target is an Application Load Balancer -- the
    #   NLB-in-front-of-ALB pattern that combines static Layer-4 IPs with
    #   Layer-7 routing.
    target_type = optional(string, "")

    # The port targets receive traffic on, 1-65535. Required for every target
    # type except "lambda". Individual target registrations can override it
    # per target. Immutable.
    port = optional(number, 0)

    # The protocol used between the load balancer and the targets. Decides the
    # load balancer family this group can attach to. Immutable.
    # - ALB: "HTTP", "HTTPS".
    # - NLB: "TCP", "UDP", "TCP_UDP", "TLS", "QUIC", "TCP_QUIC".
    # Required for every target type except "lambda". A TLS listener may
    # forward to TCP targets (the NLB terminates TLS); an HTTPS listener may
    # forward to HTTP targets (the ALB terminates TLS). "QUIC" carries QUIC
    # traffic natively; "TCP_QUIC" serves both TCP and QUIC on one group (the
    # HTTP/3 pattern where clients may fall back to TCP). QUIC targets are
    # registered with a quic_server_id (see targets).
    protocol = optional(string, "")

    # The application-layer protocol between an ALB and HTTP/HTTPS targets.
    # Valid values: "HTTP1" (default), "HTTP2", "GRPC". Immutable.
    # Choose "GRPC" for gRPC services (enables gRPC-native health-check
    # matchers) and "HTTP2" for end-to-end HTTP/2. Only meaningful when
    # protocol is HTTP or HTTPS.
    protocol_version = optional(string, "")

    # The address family of registered targets, for "instance" and "ip" target
    # types. Valid values: "ipv4" (default), "ipv6". Immutable. An IPv6 target
    # group requires a dualstack load balancer.
    ip_address_type = optional(string, "")

    # Health check configuration. When omitted, AWS applies protocol-appropriate
    # defaults (HTTP GET "/" for ALB protocols; TCP reachability on the traffic
    # port for NLB protocols).
    health_check = optional(object({
      # Whether health checks run at all. AWS default: true. Health checks can
      # only be disabled for "lambda" target groups; every other type requires
      # them. Optional rather than plain bool so that false ("disable") is
      # distinguishable from unset ("keep the AWS default of true").
      enabled = optional(bool)

      # Protocol for the health check probe: "HTTP", "HTTPS", or "TCP".
      # AWS default: "HTTP" for ALB-protocol groups, "TCP" for NLB-protocol
      # groups. TCP health checks are not allowed when the traffic protocol is
      # HTTP/HTTPS. Not applicable to "lambda" targets (always an invocation).
      protocol = optional(string, "")

      # Port for the health check probe: "traffic-port" (AWS default -- probe
      # whatever port each target receives traffic on) or a specific port number
      # as a string (e.g. "8081" for a dedicated health/admin port).
      port = optional(string, "")

      # Destination path for HTTP/HTTPS probes. AWS default: "/". For GRPC
      # protocol_version, the path is a fully-qualified gRPC method name
      # (AWS default "/AWS.ALB/healthcheck"). Not valid for TCP probes.
      path = optional(string, "")

      # Consecutive successful probes before an unhealthy target is considered
      # healthy. Range 2-10. AWS default: 5 (ALB) / 3 (NLB).
      healthy_threshold = optional(number, 0)

      # Consecutive failed probes before a healthy target is considered
      # unhealthy. Range 2-10. AWS default: 2 (ALB) / 3 (NLB).
      unhealthy_threshold = optional(number, 0)

      # Seconds between probes of an individual target. Range 5-300.
      # AWS default: 30.
      interval_seconds = optional(number, 0)

      # Seconds to wait for a probe response before counting it failed. Must be
      # smaller than interval_seconds. Range 2-120. AWS defaults vary by
      # protocol (HTTP: 5-6, TCP: 10).
      timeout_seconds = optional(number, 0)

      # Response codes that count as healthy, for HTTP/HTTPS probes.
      # HTTP matchers: a code ("200"), a range ("200-299"), or a list
      # ("200,202"). AWS default: "200".
      # GRPC matchers (protocol_version = "GRPC"): gRPC status codes, e.g. "0"
      # or "0-99". AWS default: "12" -- gRPC UNIMPLEMENTED, so a bare health
      # stub counts as healthy; set "0" to require an OK response.
      # Not valid for TCP probes (a TCP probe has no response body to match).
      matcher = optional(string, "")
    }))

    # Session stickiness. ALB supports cookie-based stickiness ("lb_cookie",
    # "app_cookie"); NLB supports flow-hash stickiness ("source_ip",
    # "source_ip_dest_ip", "source_ip_dest_ip_proto"). When omitted,
    # stickiness is disabled.
    stickiness = optional(object({
      # The stickiness mechanism. Required.
      # - "lb_cookie" (ALB): the load balancer issues and manages its own
      #   cookie (AWSALB); duration-based.
      # - "app_cookie" (ALB): the application issues the cookie named in
      #   cookie_name; the load balancer follows it.
      # - "source_ip" (NLB): affinity by client source IP.
      # - "source_ip_dest_ip" (NLB): affinity by source and destination IP --
      #   for dualstack groups where one client may arrive on both families.
      # - "source_ip_dest_ip_proto" (NLB): affinity by source IP, destination
      #   IP, and protocol -- the narrowest flow-hash affinity.
      type = string

      # Whether stickiness is active. AWS default: true (configuring the block
      # implies enabling it). Optional rather than plain bool so that false
      # ("configured but switched off") is distinguishable from unset.
      enabled = optional(bool)

      # Seconds a "lb_cookie" or "app_cookie" association lasts before the
      # client is re-balanced. Range 1-604800 (7 days). AWS default: 86400
      # (1 day). Not applicable to "source_ip".
      cookie_duration_seconds = optional(number, 0)

      # The application cookie the load balancer follows. Required for
      # "app_cookie", not valid otherwise. Must not begin with "AWSALB" (those
      # names are reserved for the load balancer's own cookies).
      cookie_name = optional(string, "")
    }))

    # Seconds the load balancer waits before completing deregistration of a
    # draining target, letting in-flight requests finish. Range 0-3600.
    # AWS default: 300. Not supported for "lambda" targets.
    deregistration_delay_seconds = optional(number, 0)

    # ALB only. Seconds during which a newly registered target receives a
    # linearly increasing share of traffic, letting caches warm before full
    # load arrives. 0 disables slow start (AWS default); otherwise 30-900.
    # Incompatible with the "least_outstanding_requests" algorithm and with
    # stickiness.
    slow_start_seconds = optional(number, 0)

    # ALB only. How the load balancer picks a target for each request.
    # Valid values: "round_robin" (default), "least_outstanding_requests",
    # "weighted_random". "least_outstanding_requests" suits uneven request
    # costs; "weighted_random" is required for anomaly mitigation.
    load_balancing_algorithm_type = optional(string, "")

    # ALB only. Automatic anomaly mitigation: the load balancer detects targets
    # returning anomalous responses and reduces their traffic share. Valid
    # values: "on", "off" (default). Requires
    # load_balancing_algorithm_type = "weighted_random".
    load_balancing_anomaly_mitigation = optional(string, "")

    # Whether traffic may cross Availability Zones on its way to targets in
    # this group, overriding the load balancer's own cross-zone setting.
    # Valid values: "true", "false", "use_load_balancer_configuration"
    # (default). A string tri-state, mirroring the AWS API: the third value
    # means "inherit from the load balancer".
    load_balancing_cross_zone_enabled = optional(string, "")

    # NLB only. Whether targets see the original client IP in the IP header.
    # AWS defaults this per target type -- enabled for "instance" targets,
    # disabled for "ip" targets -- so leaving it unset keeps the AWS default
    # for whichever type is in use (the reason this field is optional rather
    # than a plain bool: false must be distinguishable from unset).
    preserve_client_ip = optional(bool)

    # NLB only. Send the Proxy Protocol v2 header on connections to targets,
    # carrying client connection metadata (source/destination address and
    # port, VPC endpoint ID). Targets must be configured to parse the header
    # -- enabling this against an unaware backend breaks the connection.
    proxy_protocol_v2 = optional(bool, false)

    # NLB only. Terminate connections to a deregistered target when the
    # deregistration delay expires instead of waiting for the client to close
    # them. Recommended for long-lived connections (WebSocket, gRPC streams,
    # database protocols) that would otherwise pin draining targets.
    connection_termination = optional(bool, false)

    # Lambda targets only. Deliver multi-value HTTP headers and query
    # parameters to the function as arrays instead of last-value-wins strings.
    lambda_multi_value_headers_enabled = optional(bool, false)

    # Group-level health policy: DNS failover and unhealthy-state routing
    # thresholds that act on the group as a whole (ALB and NLB).
    target_group_health = optional(object({
      # DNS failover: when healthy targets drop below the threshold, the load
      # balancer's DNS stops resolving to the affected Availability Zone (or to
      # the whole load balancer), shifting clients elsewhere.
      dns_failover = optional(object({
        # Minimum number of healthy targets, as a string: a number (e.g. "2") or
        # "off" (AWS default) to disable the count criterion.
        minimum_healthy_targets_count = optional(string, "")

        # Minimum percentage of healthy targets, as a string: "1"-"100" or "off"
        # (AWS default) to disable the percentage criterion.
        minimum_healthy_targets_percentage = optional(string, "")
      }))

      # Unhealthy-state routing: when healthy targets drop below the threshold,
      # the load balancer routes to ALL targets -- including unhealthy ones --
      # on the theory that a partially working target beats a rejected request
      # during a mass failure.
      unhealthy_state_routing = optional(object({
        # Minimum number of healthy targets, 1-max. AWS default: 1.
        minimum_healthy_targets_count = optional(number, 0)

        # Minimum percentage of healthy targets, as a string: "1"-"100" or "off"
        # (AWS default) to disable the percentage criterion.
        minimum_healthy_targets_percentage = optional(string, "")
      }))
    }))

    # NLB TCP/TLS only. What happens to established connections while a target
    # is in an unhealthy state.
    target_health_state = optional(object({
      # When false, the NLB keeps established connections to a target that turns
      # unhealthy (AWS default behavior); when true, it terminates them. Keeping
      # connections suits long-lived sessions that may ride out a transient
      # health blip; terminating suits strict fail-fast backends.
      enable_unhealthy_connection_termination = optional(bool, false)

      # Seconds an unhealthy target keeps draining established connections
      # before they are terminated. Only meaningful when
      # enable_unhealthy_connection_termination is false. Range 0-360000.
      # AWS default: 0.
      unhealthy_draining_interval_seconds = optional(number, 0)
    }))

    # Static target registrations managed with the group. Most architectures
    # leave this empty -- ECS services, auto-scaling groups, and Kubernetes
    # controllers register their own targets dynamically -- but standalone
    # EC2 instances, fixed IPs, a Lambda function, or an inner ALB are
    # registered here. Registrations are folded into this kind (not a separate
    # resource) because a registration is pure glue with no referenceable
    # identity of its own.
    targets = optional(list(object({
      # What the target is, per the group's target_type:
      # - "instance": an EC2 instance ID (defaults to referencing an
      #   AwsEc2Instance's instance_id output).
      # - "ip": a literal IP address inside the VPC or a routable peered range.
      # - "lambda": the Lambda function ARN.
      # - "alb": the inner Application Load Balancer's ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      target_id = string

      # Overrides the group's port for this one target, 1-65535. Lets one group
      # spread traffic across heterogeneous ports (e.g. several containers on
      # one host).
      port = optional(number, 0)

      # For "ip" targets outside the load balancer's VPC (peered VPC,
      # on-premises): the literal string "all". Leave unset for in-VPC targets.
      availability_zone = optional(string, "")

      # QUIC / TCP_QUIC target groups only: the QUIC server ID this target
      # serves. QUIC routes established connections by connection ID rather
      # than by 5-tuple, and the server ID ties a registration to the QUIC
      # endpoint identity the target presents.
      quic_server_id = optional(string, "")
    })), [])

    # ALB only. The port the ALB Target Optimizer agent listens on, 1-65535.
    # Setting it enables Target Optimizer for this group: an agent on each
    # target reports its readiness for new requests over this port, and the
    # ALB routes accordingly. Requires the agent to be running on every
    # target -- enabling it without the agent marks targets unavailable.
    # Immutable: changing it replaces the target group.
    target_control_port = optional(number, 0)
  })
}
