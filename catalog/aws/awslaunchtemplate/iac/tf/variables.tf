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
  description = "AwsLaunchTemplate specification"
  type = object({
    # The AWS region the launch template is created in. A template is a
    # regional object: an auto-scaling group or node group can only launch
    # from a template in its own region.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # Human-readable description recorded on each template VERSION (AWS calls
    # this the version description, up to 255 characters). Use it as a
    # change-log line: "amzn2023 + IMDSv2 + gp3", "rotate AMI 2026-07".
    description = optional(string, "")

    # The Amazon Machine Image the instance boots from (e.g.
    # "ami-0abcdef1234567890"). Optional at the template level: a template
    # without an AMI is a partial blueprint whose consumer must supply one --
    # but note that auto-scaling groups require the template they reference
    # to carry an AMI, so leave this unset only for consumers that inject
    # their own image (EKS managed node groups, EC2 Fleet overrides).
    image_id = optional(string, "")

    # The EC2 instance type launched by default (e.g. "t3.small",
    # "m7g.large"). Mutually exclusive with instance_requirements -- name an
    # exact type here, or describe what you need there and let AWS pick
    # matching types. Leave both unset when the consumer supplies the type
    # (an ASG mixed-instances override, an EKS node group).
    instance_type = optional(string, "")

    # Attribute-based instance selection: instead of naming a type, describe
    # the compute you need (vCPU and memory ranges, CPU generations,
    # accelerators, price protection) and AWS resolves the matching set of
    # instance types at launch. The foundation of Spot diversification -- a
    # fleet that can draw from dozens of pools rides out capacity events that
    # would starve a single-type group. Mutually exclusive with
    # instance_type.
    instance_requirements = optional(object({
      # Required. Memory per instance, in MiB. min is required; leave max unset
      # (0) for no upper bound.
      memory_mib = object({
        # Lower bound, inclusive.
        min = optional(number, 0)

        # Upper bound, inclusive. 0 means no upper bound.
        max = optional(number, 0)
      })

      # Required. vCPUs per instance. min is required; leave max unset (0) for
      # no upper bound.
      vcpu_count = object({
        # Lower bound, inclusive.
        min = optional(number, 0)

        # Upper bound, inclusive. 0 means no upper bound.
        max = optional(number, 0)
      })

      # Allow-list of instance types or families, with wildcards ("m5.large",
      # "m5.*", "c*"). At most 400 entries. Mutually exclusive with
      # excluded_instance_types.
      allowed_instance_types = optional(list(string), [])

      # Deny-list of instance types or families, with wildcards. At most 400
      # entries. Mutually exclusive with allowed_instance_types.
      excluded_instance_types = optional(list(string), [])

      # Instance generations to include: "current" and/or "previous". AWS
      # default: any generation matching the other requirements.
      instance_generations = optional(list(string), [])

      # CPU manufacturers to include: "intel", "amd", "amazon-web-services"
      # (Graviton), "apple". AWS default: any. Selecting only
      # "amazon-web-services" is how an arm64 fleet is expressed -- pair it
      # with an arm64 AMI.
      cpu_manufacturers = optional(list(string), [])

      # Bare-metal eligibility: "included", "excluded" (AWS default), or
      # "required".
      bare_metal = optional(string, "")

      # Burstable (T-family) eligibility: "included", "excluded" (AWS
      # default), or "required".
      burstable_performance = optional(string, "")

      # Only instance types that support hibernation.
      require_hibernate_support = optional(bool, false)

      # Spot price protection: exclude types whose Spot price exceeds the
      # identified lowest-priced type's Spot price by more than this
      # percentage. Mutually exclusive with
      # max_spot_price_as_percentage_of_optimal_on_demand_price.
      spot_max_price_percentage_over_lowest_price = optional(number, 0)

      # Spot price protection anchored to On-Demand: exclude types whose Spot
      # price exceeds this percentage of the optimal type's On-Demand price --
      # steadier than the lowest-Spot anchor because On-Demand prices do not
      # fluctuate. Mutually exclusive with
      # spot_max_price_percentage_over_lowest_price.
      max_spot_price_as_percentage_of_optimal_on_demand_price = optional(number, 0)

      # On-Demand price protection: exclude types whose On-Demand price exceeds
      # the identified lowest-priced type's by more than this percentage. AWS
      # default: 20.
      on_demand_max_price_percentage_over_lowest_price = optional(number, 0)

      # Instance-store (local disk) eligibility: "included" (AWS default),
      # "excluded", or "required".
      local_storage = optional(string, "")

      # Local storage technologies when instance-store is in play: "hdd"
      # and/or "ssd".
      local_storage_types = optional(list(string), [])

      # Total local (instance-store) storage, in GB.
      total_local_storage_gb = optional(object({
        # Lower bound, inclusive.
        min = optional(number, 0)

        # Upper bound, inclusive. 0 means no upper bound.
        max = optional(number, 0)
      }))

      # Memory-to-vCPU ratio, in GiB per vCPU -- a compact way to say "memory
      # optimized" (min 8) or "compute optimized" (max 2) without naming
      # families.
      memory_gib_per_vcpu = optional(object({
        # Lower bound, inclusive.
        min = optional(number, 0)

        # Upper bound, inclusive. 0 means no upper bound.
        max = optional(number, 0)
      }))

      # Number of network interfaces the type must support.
      network_interface_count = optional(object({
        # Lower bound, inclusive.
        min = optional(number, 0)

        # Upper bound, inclusive. 0 means no upper bound.
        max = optional(number, 0)
      }))

      # Network bandwidth, in Gbps.
      network_bandwidth_gbps = optional(object({
        # Lower bound, inclusive.
        min = optional(number, 0)

        # Upper bound, inclusive. 0 means no upper bound.
        max = optional(number, 0)
      }))

      # Baseline (non-burst) EBS bandwidth, in Mbps.
      baseline_ebs_bandwidth_mbps = optional(object({
        # Lower bound, inclusive.
        min = optional(number, 0)

        # Upper bound, inclusive. 0 means no upper bound.
        max = optional(number, 0)
      }))

      # Number of accelerators (GPUs, FPGAs, inference chips). Set min 1 to
      # require accelerated types; set max 0 explicitly via {min:0, max:0} is
      # not expressible -- to EXCLUDE accelerators, leave this unset and rely
      # on accelerator_types being empty.
      accelerator_count = optional(object({
        # Lower bound, inclusive.
        min = optional(number, 0)

        # Upper bound, inclusive. 0 means no upper bound.
        max = optional(number, 0)
      }))

      # Accelerator manufacturers: "nvidia", "amd", "amazon-web-services",
      # "xilinx", "habana".
      accelerator_manufacturers = optional(list(string), [])

      # Specific accelerator models (e.g. "a100", "v100", "t4",
      # "inferentia", "radeon-pro-v520").
      accelerator_names = optional(list(string), [])

      # Accelerator categories: "gpu", "fpga", "inference".
      accelerator_types = optional(list(string), [])

      # Total accelerator memory, in MiB.
      accelerator_total_memory_mib = optional(object({
        # Lower bound, inclusive.
        min = optional(number, 0)

        # Upper bound, inclusive. 0 means no upper bound.
        max = optional(number, 0)
      }))
    }))

    # The name of an existing EC2 key pair injected for SSH access. Leave
    # unset for keyless fleets (SSM Session Manager via the instance
    # profile is the modern posture).
    key_name = optional(string, "")

    # Instance user data: a cloud-init config or shell script executed on
    # first boot. Provide PLAIN TEXT here -- both IaC modules base64-encode
    # it for the EC2 API, so the manifest stays readable. AWS limit: 16 KiB
    # before encoding.
    user_data = optional(string, "")

    # The IAM instance profile attached to launched instances -- the
    # instance's identity for SSM access, ECR pulls, S3 access, and every
    # other AWS API call the workload makes. Reference an
    # AwsIamInstanceProfile's instance_profile_arn output or pass a literal
    # profile ARN.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    instance_profile = optional(string, "")

    # Security groups attached to the instance's primary network interface.
    # Mutually exclusive with per-interface security groups inside
    # network_interfaces -- when the template declares explicit interfaces,
    # attach security groups on the interface instead.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # Dedicated EBS throughput between the instance and its volumes. Only
    # meaningful for instance types where EBS optimization is optional (most
    # current-generation types have it always-on at no charge); enabling it
    # on an unsupported type fails the launch.
    ebs_optimized = optional(bool, false)

    # Block device mappings: the volumes attached at launch, keyed by device
    # name. Override the AMI's root volume (grow it, switch to gp3, encrypt
    # with a CMK) or attach additional data volumes.
    block_device_mappings = optional(list(object({
      # The device name exposed to the instance (e.g. "/dev/xvda",
      # "/dev/sdf"). Required.
      device_name = string

      # An instance-store virtual device name ("ephemeral0", "ephemeral1", ...)
      # for instance types with local disks. Mutually exclusive with ebs.
      virtual_name = optional(string, "")

      # Suppress a device the AMI would otherwise attach -- the way to DROP an
      # AMI-baked data volume. Mutually exclusive with ebs.
      no_device = optional(bool, false)

      # EBS volume configuration for this device.
      ebs = optional(object({
        # Volume size in GiB. Must be at least the AMI snapshot's size when
        # overriding the root device.
        volume_size_gb = optional(number, 0)

        # Volume type: "gp3" (the current general-purpose default choice),
        # "gp2", "io1", "io2" (provisioned IOPS), "st1", "sc1" (throughput/cold
        # HDD), "standard" (legacy magnetic). Unset inherits from the AMI
        # mapping.
        volume_type = optional(string, "")

        # Provisioned IOPS. Required for "io1"/"io2"; optional for "gp3"
        # (baseline 3000 without it); not valid for other types.
        iops = optional(number, 0)

        # Throughput in MiB/s, 125-2000. "gp3" only (baseline 125 without it);
        # gp3 volumes support up to 2,000 MiB/s.
        throughput_mibps = optional(number, 0)

        # Encrypt the volume at rest. Snapshots and volumes created from them
        # stay encrypted. When the account enforces EBS-encryption-by-default
        # this is already true regardless.
        encrypted = optional(bool, false)

        # The KMS key for encryption. Reference an AwsKmsKey's key_arn output or
        # pass a literal key ARN. Unset with encrypted = true uses the AWS
        # managed aws/ebs key; a customer-managed key adds revocation and
        # cross-account control.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        kms_key_id = optional(string, "")

        # Create the volume from this EBS snapshot (e.g. a pre-warmed data
        # volume).
        snapshot_id = optional(string, "")

        # Delete the volume when the instance terminates. AWS default: true for
        # the root volume, false for additional volumes (inherited from the
        # AMI). Optional rather than plain bool so an explicit false ("keep the
        # data volume") is distinguishable from unset ("keep the AMI default").
        delete_on_termination = optional(bool)

        # MiB/s at which the volume hydrates from its snapshot, 100-300.
        # Without it, snapshot blocks load lazily on first read (the classic
        # cold-start latency on restored volumes); a paid initialization rate
        # makes the volume fully warm on a schedule. Only meaningful with
        # snapshot_id.
        volume_initialization_rate_mibps = optional(number, 0)
      }))
    })), [])

    # Explicit network interfaces. Most templates leave this empty and let
    # the consumer place the instance (an ASG spreads across its subnets);
    # declare interfaces to control public-IP association, static private
    # IPs/prefixes, multiple NICs, or EFA for HPC/ML workloads. When any
    # interface is declared, put security groups on the interface, not on
    # security_group_ids.
    network_interfaces = optional(list(object({
      # Position of the interface in the attachment order. The primary
      # interface is 0.
      device_index = optional(number, 0)

      # The physical network card the interface binds to, for instance types
      # with multiple cards (high-bandwidth/EFA types). Default 0.
      network_card_index = optional(number, 0)

      # Free-text description of the interface.
      description = optional(string, "")

      # Interface type: "interface" (standard, the default), "efa" (Elastic
      # Fabric Adapter with OS-bypass for tightly coupled HPC/ML), or
      # "efa-only" (EFA without an IP -- secondary cards on multi-card
      # types).
      interface_type = optional(string, "")

      # Attach an existing ENI by ID instead of creating one -- for a
      # pre-provisioned static identity (fixed IP/MAC). Mutually exclusive
      # with subnet placement and addressing fields.
      network_interface_id = optional(string, "")

      # Associate a public IPv4 address. Optional tri-state: unset inherits
      # the subnet's map-public-IP setting; an explicit value overrides it
      # either way.
      associate_public_ip_address = optional(bool)

      # Delete the interface when the instance terminates. AWS default: true
      # for interfaces the launch creates. Optional so an explicit false
      # ("keep the ENI for reuse") is distinguishable from unset.
      delete_on_termination = optional(bool)

      # The subnet the interface lives in. Reference an AwsSubnet's subnet_id
      # output or pass a literal subnet ID. Setting it pins every launch to
      # this subnet; auto-scaling templates normally leave it unset.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnet_id = optional(string, "")

      # Security groups on this interface. Reference AwsSecurityGroup
      # security_group_id outputs or pass literal group IDs.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      security_group_ids = optional(list(string), [])

      # A specific primary private IPv4 address from the subnet's range.
      private_ip_address = optional(string, "")

      # Number of additional private IPv4 addresses AWS auto-assigns.
      # Mutually exclusive with ipv4_addresses.
      ipv4_address_count = optional(number, 0)

      # Specific secondary private IPv4 addresses. Mutually exclusive with
      # ipv4_address_count.
      ipv4_addresses = optional(list(string), [])

      # Number of IPv6 addresses AWS auto-assigns from the subnet's IPv6
      # range. Mutually exclusive with ipv6_addresses.
      ipv6_address_count = optional(number, 0)

      # Specific IPv6 addresses. Mutually exclusive with ipv6_address_count.
      ipv6_addresses = optional(list(string), [])

      # Number of /28 IPv4 prefixes AWS auto-assigns (prefix delegation --
      # how Kubernetes CNIs scale pod IPs per node). Mutually exclusive with
      # ipv4_prefixes.
      ipv4_prefix_count = optional(number, 0)

      # Specific /28 IPv4 prefixes. Mutually exclusive with
      # ipv4_prefix_count.
      ipv4_prefixes = optional(list(string), [])

      # Number of /80 IPv6 prefixes AWS auto-assigns. Mutually exclusive with
      # ipv6_prefixes.
      ipv6_prefix_count = optional(number, 0)

      # Specific /80 IPv6 prefixes. Mutually exclusive with
      # ipv6_prefix_count.
      ipv6_prefixes = optional(list(string), [])

      # Associate a Carrier IP (AWS Wavelength zones) -- the Wavelength
      # analog of a public IP, reachable from the carrier's mobile
      # network. Only meaningful for interfaces in Wavelength-zone
      # subnets. Optional tri-state like associate_public_ip_address.
      associate_carrier_ip_address = optional(bool)

      # Idle-timeout tuning for the interface's connection tracking --
      # reclaim conntrack slots faster on connection-churning workloads
      # (load-balancer backends, NAT-ish proxies).
      connection_tracking = optional(object({
        # Idle timeout for established TCP connections, 60-432000 seconds
        # (5 days, the AWS default).
        tcp_established_timeout_seconds = optional(number, 0)

        # Idle timeout for UDP "connections" that have seen traffic BOTH
        # ways (streams), 60-180 seconds.
        udp_stream_timeout_seconds = optional(number, 0)

        # Idle timeout for single-direction UDP flows, 30-60 seconds.
        udp_timeout_seconds = optional(number, 0)
      }))

      # ENA Express (SRD -- Scalable Reliable Datagram): reduce tail
      # latency between instances that both enable it, on supported
      # types. TCP benefits transparently; UDP needs the explicit flag.
      ena_srd = optional(object({
        # Turn ENA Express on for the interface (TCP flows benefit
        # transparently).
        enabled = optional(bool, false)

        # Also run eligible UDP traffic over SRD. Only meaningful when
        # enabled is true.
        udp_enabled = optional(bool, false)
      }))

      # Number of ENA queues for the interface, on instance types that
      # support ENA queue tuning. 0 keeps the AWS default per instance
      # size.
      ena_queue_count = optional(number, 0)

      # Make the auto-assigned IPv6 address the instance's PRIMARY,
      # stable identity: it survives network-interface replacement --
      # required for IPv6-only instances whose address must not change.
      primary_ipv6 = optional(bool)
    })), [])

    # Instance Metadata Service posture. Set http_tokens = "required" to
    # enforce IMDSv2 -- the single most effective hardening against
    # credential-stealing SSRF attacks, and the recommended default for every
    # new template.
    metadata_options = optional(object({
      # Whether the metadata service is reachable at all: "enabled" (AWS
      # default) or "disabled" (nothing on the instance can fetch credentials
      # -- rare, for fully static workloads).
      http_endpoint = optional(string, "")

      # IMDS version enforcement: "required" (IMDSv2 session tokens only --
      # the recommended hardening) or "optional" (v1 and v2 both answer; the
      # AWS default for backward compatibility).
      http_tokens = optional(string, "")

      # TTL of the IMDSv2 token-fetch packet, 1-64. 1 (AWS default) confines
      # metadata to the instance itself; 2 lets containerized workloads (one
      # extra network hop) reach it.
      http_put_response_hop_limit = optional(number, 0)

      # Serve metadata over the interface's IPv6 endpoint as well: "enabled"
      # or "disabled" (AWS default).
      http_protocol_ipv6 = optional(string, "")

      # Expose the instance's tags through the metadata service: "enabled" or
      # "disabled" (AWS default). Lets on-instance agents read tags without
      # ec2:DescribeTags permission.
      instance_metadata_tags = optional(string, "")
    }))

    # Detailed CloudWatch monitoring: metrics at 1-minute granularity instead
    # of the free 5-minute default. Costs per instance-metric; auto-scaling
    # policies react meaningfully faster with it enabled.
    detailed_monitoring = optional(bool, false)

    # Placement of launched instances: availability zone pinning, placement
    # group membership, and tenancy.
    placement = optional(object({
      # Pin launches to one availability zone (e.g. "us-west-2a"). For ASG
      # templates, prefer the group's subnets over AZ pinning here.
      availability_zone = optional(string, "")

      # The placement group to launch into: "cluster" groups pack instances
      # for lowest latency, "spread" and "partition" groups separate them for
      # fault isolation. Mutually exclusive with group_id.
      group_name = optional(string, "")

      # Partition number within a partition placement group.
      partition_number = optional(number, 0)

      # Instance tenancy: "default" (shared hardware), "dedicated"
      # (single-tenant hardware), or "host" (a specific Dedicated Host --
      # licensing scenarios).
      tenancy = optional(string, "")

      # The placement group by ID ("pg-...") instead of name. Mutually
      # exclusive with group_name.
      group_id = optional(string, "")

      # The Dedicated Host to land on ("h-..."). Host tenancy pins BYOL
      # workloads (per-socket/per-core licenses) to specific hardware.
      # Mutually exclusive with host_resource_group_arn.
      host_id = optional(string, "")

      # A License Manager host resource group ARN -- AWS picks (and
      # manages) a Dedicated Host from the group instead of naming one.
      # Mutually exclusive with host_id.
      host_resource_group_arn = optional(string, "")

      # Dedicated Host affinity: "default" (instance may restart on any
      # host) or "host" (the instance always restarts on the same host --
      # needed when licenses are bound to specific hardware).
      affinity = optional(string, "")

      # Named spread domain on AWS Outposts placement groups.
      spread_domain = optional(string, "")
    }))

    # CPU topology and features: trim vCPUs on license-bound workloads
    # (threads_per_core = 1), or enable AMD SEV-SNP memory encryption.
    cpu_options = optional(object({
      # Number of physical cores. Combined with threads_per_core = 1 this
      # trims the vCPU count -- the standard move for per-core licensed
      # software (databases) on large instance types.
      core_count = optional(number, 0)

      # Threads per core: 2 (hyper-threading, the default on x86) or 1
      # (disable SMT -- per-core licensing, or HPC codes that fight over
      # shared core resources).
      threads_per_core = optional(number, 0)

      # AMD SEV-SNP memory encryption on supported AMD types: "enabled" or
      # "disabled". Confidential-computing hardening.
      amd_sev_snp = optional(string, "")

      # Hardware-assisted nested virtualization on supported types:
      # "enabled" (run hypervisors/VMs inside the instance) or "disabled".
      nested_virtualization = optional(string, "")
    }))

    # Credit option for burstable (T-family) instance types: "standard"
    # (throttle when credits run out) or "unlimited" (keep bursting, pay for
    # the excess). AWS default: "unlimited" for recent T families. Ignored
    # for non-burstable types.
    cpu_credits = optional(string, "")

    # Request Spot capacity instead of On-Demand. Configuring this block
    # makes every launch from the template a Spot request -- appropriate for
    # interruption-tolerant workloads. For a mixed On-Demand/Spot fleet,
    # leave this unset and blend purchase options in the auto-scaling group's
    # mixed-instances policy instead.
    spot_options = optional(object({
      # Maximum price per instance-hour, as a decimal string (e.g. "0.05").
      # AWS default: the On-Demand price -- and leaving it unset is the AWS
      # recommendation, since Spot's discount comes from interruption risk,
      # not bidding.
      max_price = optional(string, "")

      # Spot request type: "one-time" (the default for templates consumed by
      # auto-scaling -- the ASG replaces interrupted capacity itself) or
      # "persistent" (EC2 re-requests the instance after interruption;
      # standalone instances only).
      spot_instance_type = optional(string, "")

      # What happens to the instance on interruption: "terminate" (default),
      # "stop", or "hibernate". "stop"/"hibernate" require a persistent
      # request.
      instance_interruption_behavior = optional(string, "")

      # Expiry of a persistent request, RFC3339 (e.g. "2027-01-01T00:00:00Z").
      # Only valid when spot_instance_type is "persistent".
      valid_until = optional(string, "")
    }))

    # Launch instances as AWS Nitro Enclaves parents, enabling isolated
    # enclave VMs for secret-processing workloads. Not supported on every
    # instance type; incompatible with hibernation.
    enclave_enabled = optional(bool, false)

    # Pre-provision instances for hibernation (encrypted root volume large
    # enough to hold RAM contents required). Incompatible with Nitro
    # Enclaves.
    hibernation_enabled = optional(bool, false)

    # Simplified automatic recovery on instance impairment: "default"
    # (recover onto healthy hardware) or "disabled" (leave the instance
    # impaired -- for workloads with instance-store state that recovery would
    # silently discard).
    auto_recovery = optional(string, "")

    # How the guest's private DNS hostname is formed and which DNS records
    # resolve to it.
    private_dns_name_options = optional(object({
      # Hostname scheme: "ip-name" (ip-10-0-1-5.ec2.internal, the classic
      # default) or "resource-name" (i-0123....ec2.internal -- stable across
      # IP changes, required for IPv6-only subnets).
      hostname_type = optional(string, "")

      # Publish an A record (IPv4) for the hostname.
      enable_resource_name_dns_a_record = optional(bool, false)

      # Publish an AAAA record (IPv6) for the hostname.
      enable_resource_name_dns_aaaa_record = optional(bool, false)
    }))

    # Protect launched instances from being stopped via the API. A guard for
    # pet-like fleet members; leave false for disposable fleet instances.
    disable_api_stop = optional(bool, false)

    # Protect launched instances from being terminated via the API. NOTE: an
    # auto-scaling group cannot scale in instances with termination
    # protection -- use the ASG's protect_from_scale_in instead for fleet
    # instances.
    disable_api_termination = optional(bool, false)

    # What an OS-initiated shutdown does to the instance: "stop" (AWS
    # default; the instance can be started again) or "terminate" (the
    # instance is gone -- the right choice for immutable fleet members that
    # should never be resurrected by hand).
    instance_initiated_shutdown_behavior = optional(string, "")

    # Launch into EC2 Capacity Reservations -- guaranteed capacity a
    # reserved fleet has already paid for (On-Demand Capacity
    # Reservations, or Capacity Blocks for ML). Leave unset for the AWS
    # default (use a matching open reservation when one exists).
    capacity_reservation = optional(object({
      # How launches relate to reservations:
      # - "open": use a matching open reservation when one exists (the AWS
      #   default behavior), otherwise launch as regular On-Demand.
      # - "none": never consume a reservation, even a matching open one.
      # - "capacity-reservations-only": launch ONLY into reserved capacity
      #   -- fail rather than fall back to On-Demand.
      # Leave empty when targeting a specific reservation below.
      preference = optional(string, "")

      # A specific Capacity Reservation to launch into ("cr-..."). The
      # required target for Capacity Blocks. Mutually exclusive with
      # capacity_reservation_resource_group_arn.
      capacity_reservation_id = optional(string, "")

      # A resource-group ARN collecting Capacity Reservations -- target the
      # group instead of one reservation ID. Mutually exclusive with
      # capacity_reservation_id.
      capacity_reservation_resource_group_arn = optional(string, "")
    }))

    # The purchase market for launched instances: "spot" (pairs with
    # spot_options) or "capacity-block" (pre-purchased Capacity Blocks
    # for ML -- requires capacity_reservation targeting the block).
    # Unset means On-Demand, unless spot_options implies "spot".
    market_type = optional(string, "")

    # Network bandwidth weighting on supported types: "default", "vpc-1"
    # (shift baseline bandwidth toward VPC networking), or "ebs-1"
    # (shift it toward EBS throughput). A no-cost performance bias for
    # network- or storage-heavy workloads.
    bandwidth_weighting = optional(string, "")

    # AWS License Manager license-configuration ARNs launched instances
    # consume -- BYOL license tracking (e.g. per-core database
    # licenses). Literal ARNs: License Manager has no Planton kind.
    license_configuration_arns = optional(list(string), [])

    # Additional network interfaces attached at launch BEYOND the
    # primary set in network_interfaces -- a 2026 EC2 capability for
    # multi-homed instances (e.g. a dedicated storage-subnet interface).
    # Interfaces are created and deleted with the instance.
    secondary_interfaces = optional(list(object({
      # The device index the interface attaches at.
      device_index = optional(number, 0)

      # The network card the interface binds to, on instance types with
      # multiple network cards.
      network_card_index = optional(number, 0)

      # Delete the interface when the instance terminates.
      delete_on_termination = optional(bool, false)

      # The subnet the secondary interface lives in. Reference an
      # AwsSubnet's subnet_id output or pass a literal subnet ID -- e.g. a
      # dedicated storage or replication subnet.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      secondary_subnet_id = optional(string, "")

      # Number of private IPv4 addresses AWS auto-assigns on the
      # interface. Mutually exclusive with private_ip_addresses.
      private_ip_address_count = optional(number, 0)

      # Specific private IPv4 addresses to assign. Mutually exclusive with
      # private_ip_address_count.
      private_ip_addresses = optional(list(string), [])
    })), [])
  })
}
