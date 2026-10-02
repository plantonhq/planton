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
  description = "GcpPscServiceAttachment specification"
  type = object({
    # The GCP project that owns the service attachment (the producer
    # project). A literal project ID or a reference to a GcpProject. If
    # omitted, the provider's default project is used. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the service attachment in GCP. 1-63 characters, lowercase
    # letters, digits, and hyphens, starting with a letter and not ending
    # with a hyphen. Defaults to metadata.name. Immutable.
    attachment_name = optional(string, "")

    # The region of the attachment -- the same region as the target
    # forwarding rule and the NAT subnets. Required. Immutable.
    region = string

    # Human-readable description of the attachment.
    description = optional(string, "")

    # The producer's load balancer: the regional forwarding rule (a
    # GcpGlobalForwardingRule with `region` set) of the internal passthrough
    # Network Load Balancer or internal Application Load Balancer that
    # serves the published service, in this attachment's region. Required;
    # changing it moves the published service in place.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    target_service = string

    # The PSC NAT subnets consumer traffic is translated into: one or more
    # GcpSubnetwork of purpose PRIVATE_SERVICE_CONNECT in the producer VPC
    # and this region. Required, at least one. Size them for the consumer
    # endpoint count; add subnets to grow capacity. The provider treats the
    # list as a set, so order never diffs.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    nat_subnets = list(string)

    # Who may connect:
    # - ACCEPT_AUTOMATIC: any consumer in any project connects without
    #   approval (subject to the reject list)
    # - ACCEPT_MANUAL: only consumers in consumer_accept_lists connect; every
    #   other connection request stays PENDING until accepted
    # Required. Mutable.
    connection_preference = string

    # Consumers admitted under ACCEPT_MANUAL, each with its connection limit.
    # The provider compares the list as a set.
    consumer_accept_lists = optional(list(object({
      # The consumer project allowed to connect: a project ID or number, or a
      # reference to a GcpProject. Every network in the project may connect.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      project_id = optional(string, "")

      # The consumer VPC network allowed to connect (a network self-link or a
      # reference to a GcpVpcNetwork); narrower than a project.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      network = optional(string, "")

      # The specific consumer endpoint (a consumer forwarding rule's URL)
      # allowed to connect; the narrowest form.
      endpoint_url = optional(string, "")

      # How many Private Service Connect endpoints this consumer may connect to
      # the attachment at once. Required, at least 1.
      connection_limit = number
    })), [])

    # Consumer projects (IDs or numbers, or GcpProject references) refused
    # even under ACCEPT_AUTOMATIC. Compared as a set.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    consumer_reject_lists = optional(list(string), [])

    # Whether a change to the accept or reject lists reconciles EXISTING
    # connections: false (Google's default) only affects PENDING endpoints,
    # so an accepted consumer stays connected after being removed from the
    # accept list; true moves existing ACCEPTED endpoints to REJECTED when
    # their project lands on the reject list, and vice versa. Unset lets
    # Google apply its default.
    reconcile_connections = optional(bool)

    # Enable the PROXY protocol on connections through this attachment, so
    # the producer's backends receive the consumer's original TCP/IP address
    # data in the connection header. Required by the API (state it either
    # way); the backends must speak the protocol when true.
    enable_proxy_protocol = optional(bool, false)

    # Domain name registered with Cloud DNS for the connected endpoints, e.g.
    # "p.mycompany.com." (with the trailing dot). At most one. Immutable.
    domain_names = optional(list(string), [])

    # How many consumer spokes a connected PSC endpoint may be propagated to
    # through Network Connectivity Center; per accept-list entry under
    # ACCEPT_MANUAL, per consumer project under ACCEPT_AUTOMATIC. Unset lets
    # Google apply its default (250); an explicit 0 is sent as 0.
    propagated_connection_limit = optional(number)

    # Show the NAT IP addresses of every connected endpoint in the
    # attachment's connected-endpoints listing. Google's API currently
    # ignores the flag (the provider records the value it sends); modeled so
    # the manifest states the intent it will honor once the API does.
    show_nat_ips = optional(bool, false)

    # What destroy does to the attachment:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the attachment is deleted; every consumer endpoint
    #                connected through it loses the service
    #   "PREVENT" -- destroy FAILS; protects a published service consumers
    #                depend on
    #   "ABANDON" -- the attachment leaves management but keeps serving
    deletion_policy = optional(string, "")
  })
}
