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
  description = "AwsElasticIp specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # EC2 IPv4 address pool identifier to allocate from. Set this to a BYOIP pool
    # ID when you want the EIP to come from your own registered IP address range
    # instead of Amazon's pool. When omitted, Amazon allocates from its own pool.
    #
    # This field is ForceNew: changing it requires replacing the EIP.
    public_ipv4_pool = optional(string, "")

    # Request a specific IP address from the BYOIP pool identified by
    # `public_ipv4_pool`. The address must belong to the specified pool.
    # Only meaningful when `public_ipv4_pool` is set.
    #
    # This field is ForceNew: changing it requires replacing the EIP.
    address = optional(string, "")

    # Network border group that controls the location scope of this EIP. Used for
    # allocating EIPs in AWS Local Zones or Wavelength zones. When omitted, the
    # EIP is scoped to the Region.
    #
    # Example values: "us-east-1", "us-east-1-wl1-bos-wlz-1" (Wavelength),
    # "us-west-2-lax-1a" (Local Zone).
    #
    # This field is ForceNew: changing it requires replacing the EIP.
    network_border_group = optional(string, "")

    # Amazon VPC IP Address Manager (IPAM) pool to allocate this EIP from, e.g.
    # "ipam-pool-07ccc86aa41bef7ce". IPAM pools let network teams plan, track,
    # and audit public IPv4 usage centrally; an EIP allocated from a pool is
    # recorded against that pool's allocations. The pool must be a public-scope
    # pool provisioned for Elastic IP allocation in this region. May be combined
    # with `address` to recover a specific address the pool holds.
    #
    # Takes a literal pool id today; when the platform's IPAM component lands,
    # reference its pool output instead.
    #
    # This field is ForceNew: changing it requires replacing the EIP.
    ipam_pool_id = optional(string, "")

    # EC2 instance to associate this EIP with. Reference an AwsEc2Instance's
    # instance_id output or pass a literal "i-..." id. The address follows the
    # instance across stop/start cycles until disassociated. At most one of
    # `instance` and `network_interface` may be set — associating with an
    # instance targets its primary network interface's primary private IP.
    # Updates in place: changing the target re-associates the same address
    # (the allocation is never replaced).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    instance = optional(string, "")

    # Network interface (ENI) to associate this EIP with — the precise form of
    # association: pick the exact ENI (and, with associate_with_private_ip, the
    # exact private address) the public IP maps to. Reference an
    # AwsEc2Instance's primary_network_interface_id output or pass a literal
    # "eni-..." id (standalone ENIs, appliance interfaces). At most one of
    # `instance` and `network_interface` may be set. Updates in place.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    network_interface = optional(string, "")

    # The specific private IPv4 address on the association target that this
    # EIP maps to. Only meaningful when the target (instance or ENI) carries
    # multiple private addresses — omitted, AWS uses the primary private IP.
    # Requires `instance` or `network_interface` to be set. Updates in place.
    associate_with_private_ip = optional(string, "")

    # Fully qualified domain name to set as this EIP's reverse DNS (PTR)
    # record, e.g. "mail.example.com" — required by many mail providers before
    # they accept SMTP traffic from the address. AWS validates SERVER-SIDE that
    # a forward DNS record (A) for this domain already resolves to the EIP's
    # address BEFORE granting the PTR record, so create the EIP first, point
    # the domain at its public_ip output, then set this field on a follow-up
    # apply. Updates in place; clearing the field resets the PTR record.
    reverse_dns_domain_name = optional(string, "")
  })
}
