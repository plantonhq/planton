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
  description = "GcpNetworkEndpointGroup specification"
  type = object({
    # The GCP project that owns the group. A literal project ID or a
    # reference to a GcpProject. If omitted, the provider's default project
    # is used. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the group in GCP. 1-63 characters, lowercase letters, digits,
    # and hyphens, starting with a letter and not ending with a hyphen.
    # Defaults to metadata.name. Immutable.
    neg_name = optional(string, "")

    # The scope selector. Empty builds a GLOBAL internet network endpoint
    # group (INTERNET_IP_PORT or INTERNET_FQDN_PORT, for a global external
    # Application Load Balancer); a zone name such as us-central1-a builds a
    # ZONAL group in that zone (VM, hybrid, or internet endpoints, for
    # regional and passthrough load balancers). A zonal group needs a
    # network; a global one has none. Immutable: a group cannot move
    # between scopes or zones.
    zone = optional(string, "")

    # Human-readable description of the group. Immutable.
    description = optional(string, "")

    # The kind of endpoint every member is. Zonal: GCE_VM_IP_PORT (default;
    # VM IP and port -- Application and proxy load balancers), GCE_VM_IP
    # (VM IP only -- passthrough Network Load Balancers), NON_GCP_PRIVATE_IP_PORT
    # (hybrid: on-premises or other-cloud addresses reached over VPN or
    # Interconnect; only backend services with the EXTERNAL, EXTERNAL_MANAGED,
    # INTERNAL_MANAGED, or INTERNAL_SELF_MANAGED scheme in RATE or CONNECTION
    # mode), INTERNET_IP_PORT, INTERNET_FQDN_PORT (internet endpoints for a
    # regional external ALB), GCE_VM_IP_DEDICATED_BACKEND. Global:
    # INTERNET_IP_PORT or INTERNET_FQDN_PORT (required). Immutable.
    network_endpoint_type = optional(string, "")

    # The VPC network every endpoint belongs to (zonal groups; required
    # there): a network self-link or a reference to a GcpVpcNetwork.
    # Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    network = optional(string, "")

    # The subnetwork every endpoint belongs to (zonal groups; optional): a
    # subnetwork self-link or a reference to a GcpSubnetwork in the group's
    # region. VM endpoints' IPs must fall inside it. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnetwork = optional(string, "")

    # The port an endpoint listens on when its own port is unset. Unset
    # sends nothing (Google records none). Not for GCE_VM_IP groups.
    # Immutable.
    default_port = optional(number)

    # The group's members. The list is the whole membership: an endpoint
    # removed here is detached from the group in place, an endpoint added is
    # attached. Leave empty to create the group and attach endpoints later
    # (an autoscaler or another controller may own membership).
    endpoints = optional(list(object({
      # The VM the endpoint belongs to: a reference to a GcpComputeInstance or
      # its instance name in the group's zone. Zonal GCE_VM_IP and
      # GCE_VM_IP_PORT groups only.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      instance = optional(string, "")

      # The endpoint's IP address. For VM endpoints a primary or alias IP of
      # the instance in the group's subnetwork (unset means the instance's
      # primary internal IP); for hybrid and internet endpoints the reachable
      # address, required.
      ip_address = optional(string, "")

      # Fully qualified domain name of an INTERNET_FQDN_PORT endpoint
      # (resolved by Google at connection time).
      fqdn = optional(string, "")

      # Port the endpoint listens on. Unset falls back to the group's
      # default_port; GCE_VM_IP groups carry no port at all.
      port = optional(number)
    })), [])

    # What destroy does to the group and its endpoints:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the group is deleted (GCP refuses while a backend
    #                service still uses it)
    #   "PREVENT" -- destroy FAILS
    #   "ABANDON" -- the group leaves management but keeps serving
    deletion_policy = optional(string, "")
  })
}
