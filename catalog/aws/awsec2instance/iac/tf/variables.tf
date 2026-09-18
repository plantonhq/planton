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
  description = "AwsEc2Instance specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # The Amazon Machine Image the instance boots from (e.g.
    # "ami-0abcdef1234567890"). AMI IDs are region-specific. Required unless
    # launch_template supplies one. ForceNew: changing the AMI replaces the
    # instance (see user_data_replace_on_change for the related user-data
    # semantics).
    ami = optional(string, "")

    # The EC2 instance type (e.g. "t4g.nano", "m7g.large", "c5.xlarge")
    # determining vCPU count, memory, and network/EBS bandwidth. Required
    # unless launch_template supplies one. Changing the type stops and
    # restarts the instance in place (EBS-backed instances only) -- UNLESS
    # the old and new types share no CPU architecture (x86 -> ARM), which
    # replaces the instance.
    instance_type = optional(string, "")

    # Launch from an AwsLaunchTemplate instead of (or in addition to) the
    # inline fields. Every inline field set here OVERRIDES the template's
    # value for this instance -- the template is the org's golden baseline,
    # the inline fields are this pet's deviations.
    launch_template = optional(object({
      # The launch template ID. Reference an AwsLaunchTemplate's
      # launch_template_id output or pass a literal ID (e.g.
      # "lt-0123456789abcdef0"). Mutually exclusive with name.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      id = optional(string, "")

      # The launch template name, for templates managed outside the resource
      # graph. Mutually exclusive with id.
      name = optional(string, "")

      # The template version to launch from: a version number ("12"),
      # "$Latest" (track every new version -- each publish restarts the
      # instance), or "$Default" (AWS default when unset; both Planton launch
      # template modules promote each new version to default, so this tracks
      # the template's releases). Changing this to a version the instance is
      # not already running REPLACES the instance -- the provider verifies
      # against the live instance's actual template version.
      version = optional(string, "")
    }))

    # The IAM instance profile attached to the instance -- its identity for
    # SSM Session Manager, ECR pulls, S3 access, and every other AWS API call
    # the workload makes. Reference an AwsIamInstanceProfile's
    # instance_profile_name output or pass a literal profile NAME (the EC2
    # instance API takes the profile by name, unlike launch templates which
    # accept an ARN). Attachable and replaceable in place on a running
    # instance.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    instance_profile = optional(string, "")

    # The name of an existing EC2 key pair injected for SSH access. Leave
    # unset for keyless instances (SSM Session Manager via the instance
    # profile is the modern posture). ForceNew: changing the key pair
    # replaces the instance.
    key_name = optional(string, "")

    # The subnet the instance lives in. Reference an AwsSubnet's subnet_id
    # output or pass a literal subnet ID. Unset launches into the account's
    # default VPC (its default subnet in the chosen AZ) -- acceptable for
    # experiments, not a production posture; org accounts frequently have no
    # default VPC at all. ForceNew: changing the subnet replaces the
    # instance.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_id = optional(string, "")

    # Security groups attached to the instance's primary network interface --
    # what can reach the instance and what it can reach. Reference
    # AwsSecurityGroup security_group_id outputs or pass literal group IDs.
    # Unset falls back to the VPC's default security group -- not a
    # production posture. Updatable in place.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # Attach an EXISTING network interface (ENI) as the primary interface
    # (eth0) instead of creating one -- how an instance inherits a
    # pre-provisioned static identity (fixed private IP, fixed MAC, existing
    # security groups). The ENI carries the network configuration, so the
    # subnet/security-group/addressing fields here must stay unset.
    # ForceNew: changing the ENI replaces the instance.
    primary_network_interface_id = optional(string, "")

    # A specific primary private IPv4 address from the subnet's range.
    # Unset lets AWS pick one. ForceNew: changing it replaces the instance.
    private_ip = optional(string, "")

    # Additional private IPv4 addresses on the primary interface -- for
    # hosting multiple TLS endpoints or failover addresses on one instance.
    # Updatable in place.
    secondary_private_ips = optional(list(string), [])

    # Associate a public IPv4 address at launch. Optional tri-state: unset
    # inherits the subnet's map-public-IP-on-launch setting; an explicit
    # value overrides it either way. ForceNew: changing it replaces the
    # instance. Note AWS now bills every public IPv4 address.
    associate_public_ip_address = optional(bool)

    # Source/destination checking on the primary interface. Optional
    # tri-state: unset keeps AWS's default (true -- checking on); an
    # explicit false is the forwarding posture for instances that carry
    # traffic they neither originated nor terminate (NAT instances,
    # software routers, VPN appliances) -- without it the network silently
    # drops their forwarded packets. Leave unset when attaching a
    # pre-provisioned primary ENI: the provider rejects the combination
    # (the ENI carries its own source/dest-check setting), and setting the
    # field on the ENI's kind is the right home for that posture.
    source_dest_check = optional(bool)

    # Number of IPv6 addresses AWS auto-assigns from the subnet's IPv6
    # range. Mutually exclusive with ipv6_addresses.
    ipv6_address_count = optional(number, 0)

    # Specific IPv6 addresses from the subnet's range. Mutually exclusive
    # with ipv6_address_count.
    ipv6_addresses = optional(list(string), [])

    # Designate the first IPv6 address as the instance's stable primary
    # IPv6 address (kept until the instance or ENI is deleted) -- required
    # posture for IPv6-only workloads that must keep one consistent
    # address. One-way: disabling it after enablement forces replacement.
    enable_primary_ipv6 = optional(bool)

    # How the guest's private DNS hostname is formed and which DNS records
    # resolve to it. Defaults inherit from the subnet's settings.
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

    # Additional network interfaces created and attached at launch on OTHER
    # network cards -- for high-bandwidth instance types with multiple
    # network cards. For ordinary multi-homing on card 0, create standalone
    # ENIs instead; for the primary interface, see
    # primary_network_interface_id.
    secondary_network_interfaces = optional(list(object({
      # The physical network card the interface binds to. Required; card 0
      # carries the primary interface, so secondary interfaces target 1+.
      network_card_index = optional(number, 0)

      # Position of the interface in the attachment order on its card.
      # AWS default: 0.
      device_index = optional(number, 0)

      # The subnet the interface lives in. Reference an AwsSubnet's subnet_id
      # output or pass a literal subnet ID; must be in the same AZ as the
      # instance. Security groups are not configurable on secondary
      # interfaces at launch (the EC2 API applies the VPC default group);
      # manage them post-launch on the ENI itself.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnet_id = string

      # Number of private IPv4 addresses to assign. AWS default: 1.
      private_ip_address_count = optional(number, 0)

      # Delete the interface when the instance terminates. AWS default: true.
      # Optional so an explicit false ("keep the ENI for reuse") is
      # distinguishable from unset.
      delete_on_termination = optional(bool)
    })), [])

    # Reshape the root (boot) volume the AMI defines: grow it, switch it to
    # gp3, encrypt it with a customer-managed key. Unset fields inherit
    # from the AMI's block device mapping. Size, type, IOPS, throughput,
    # and delete-on-termination are updatable in place; flipping encryption
    # requires replacement.
    root_block_device = optional(object({
      # Volume size in GiB. Must be at least the AMI snapshot's size.
      # Growable in place; shrinking requires replacement.
      volume_size_gb = optional(number, 0)

      # Volume type: "gp3" (the current general-purpose default choice),
      # "gp2", "io1", "io2" (provisioned IOPS), "st1", "sc1"
      # (throughput/cold HDD), "standard" (legacy magnetic). Unset inherits
      # from the AMI mapping.
      volume_type = optional(string, "")

      # Provisioned IOPS. Required for "io1"/"io2"; optional for "gp3"
      # (baseline 3000 without it); not valid for other types.
      iops = optional(number, 0)

      # Throughput in MiB/s, 125-2000. "gp3" only (baseline 125 without it;
      # above 1000 requires a matching iops floor per the gp3 ratio rules AWS
      # enforces at the API).
      throughput_mibps = optional(number, 0)

      # Encrypt the root volume at rest. ForceNew on the root device: flipping
      # encryption replaces the instance. When the account enforces
      # EBS-encryption-by-default this is already true regardless.
      encrypted = optional(bool, false)

      # The KMS key for encryption. Reference an AwsKmsKey's key_arn output
      # or pass a literal key ARN. Unset with encrypted = true uses the AWS
      # managed aws/ebs key; a customer-managed key adds revocation and
      # cross-account control.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_id = optional(string, "")

      # Delete the root volume when the instance terminates. AWS default:
      # true. Optional so an explicit false ("keep the boot disk for
      # forensics/reuse") is distinguishable from unset.
      delete_on_termination = optional(bool)

      # Tags on the root volume itself. Applied AFTER instance creation by a
      # separate tagging call -- incompatible with ABAC/SCP policies that
      # require tags at creation time (use the spec-level volume_tags for
      # those). Mutually exclusive with volume_tags. Updatable in place.
      tags = optional(map(string), {})
    }))

    # Additional EBS data volumes attached at launch, keyed by device name
    # (e.g. "/dev/sdf"). Each volume's shape is create-time: changing a
    # mapping replaces the instance -- attach post-launch volumes as
    # separate resources when independent lifecycles matter.
    ebs_block_devices = optional(list(object({
      # The device name exposed to the instance (e.g. "/dev/sdf",
      # "/dev/xvdb"). Required; must not collide with the root device.
      device_name = string

      # Volume size in GiB. Required unless snapshot_id supplies the size.
      volume_size_gb = optional(number, 0)

      # Volume type: "gp3", "gp2", "io1", "io2", "st1", "sc1", "standard".
      # AWS default: gp3 for new volumes.
      volume_type = optional(string, "")

      # Provisioned IOPS. Required for "io1"/"io2"; optional for "gp3".
      iops = optional(number, 0)

      # Throughput in MiB/s, 125-2000. "gp3" only.
      throughput_mibps = optional(number, 0)

      # Encrypt the volume at rest.
      encrypted = optional(bool, false)

      # The KMS key for encryption. Reference an AwsKmsKey's key_arn output
      # or pass a literal key ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_id = optional(string, "")

      # Create the volume from this EBS snapshot (e.g. a pre-warmed data
      # volume) instead of empty.
      snapshot_id = optional(string, "")

      # Delete the volume when the instance terminates. AWS default: true.
      # Set an explicit false to keep the data volume after termination.
      delete_on_termination = optional(bool)

      # Tags on this volume. Applied AFTER instance creation by a separate
      # tagging call -- incompatible with ABAC/SCP policies that require tags
      # at creation time (use the spec-level volume_tags for those).
      # Mutually exclusive with volume_tags. Updatable in place.
      tags = optional(map(string), {})
    })), [])

    # Instance-store (ephemeral local disk) mappings for instance types
    # with local disks. Data on instance store does not survive stop or
    # termination. ForceNew: changing mappings replaces the instance.
    ephemeral_block_devices = optional(list(object({
      # The device name exposed to the instance (e.g. "/dev/sdc"). Required.
      device_name = string

      # An instance-store virtual device name ("ephemeral0", "ephemeral1",
      # ...). Mutually exclusive with no_device.
      virtual_name = optional(string, "")

      # Suppress a device the AMI would otherwise attach -- the way to DROP
      # an AMI-baked mapping. Mutually exclusive with virtual_name.
      no_device = optional(bool, false)
    })), [])

    # Tags applied uniformly to EVERY EBS volume at instance creation --
    # including volumes the AMI's block-device mapping creates that are not
    # declared here. Because they ride the launch call itself, these satisfy
    # ABAC/SCP policies that require tags at creation time. Mutually
    # exclusive with per-device tags (root_block_device.tags /
    # ebs_block_devices[].tags), which allow per-volume values but are
    # applied AFTER creation by a separate tagging call. Updatable in place.
    volume_tags = optional(map(string), {})

    # Dedicated EBS throughput between the instance and its volumes. Only
    # meaningful for instance types where EBS optimization is optional
    # (most current-generation types have it always-on at no charge);
    # enabling it on an unsupported type fails the launch. ForceNew.
    ebs_optimized = optional(bool, false)

    # Instance Metadata Service posture. Set http_tokens = "required" to
    # enforce IMDSv2 -- the single most effective hardening against
    # credential-stealing SSRF attacks, and the recommended default for
    # every new instance. Updatable in place.
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

    # Detailed CloudWatch monitoring: metrics at 1-minute granularity
    # instead of the free 5-minute default. Costs per instance-metric;
    # alarms and dashboards react meaningfully faster with it enabled.
    detailed_monitoring = optional(bool, false)

    # CPU topology and features: trim vCPUs on license-bound workloads
    # (threads_per_core = 1), enable AMD SEV-SNP memory encryption, or
    # enable nested virtualization on supported types. ForceNew: CPU
    # options are fixed at launch.
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

      # Nested virtualization on supported 8th-generation Intel types
      # ("enabled" or "disabled") -- run hypervisors inside the instance.
      # Enabling it automatically disables Virtual Secure Mode.
      nested_virtualization = optional(string, "")
    }))

    # Credit option for burstable (T-family) instance types: "standard"
    # (throttle when credits run out) or "unlimited" (keep bursting, pay
    # for the excess). AWS default: "unlimited" for recent T families.
    # Ignored for non-burstable types. Updatable in place.
    cpu_credits = optional(string, "")

    # The instance's purchase market. Unset = On-Demand (with spot_options
    # present implying "spot" for that classic shape). "spot" pairs with
    # spot_options; "capacity-block" launches into a pre-purchased ML
    # Capacity Block (target the block's reservation via
    # capacity_reservation -- required); "interruptible-capacity-reservation"
    # launches into an interruptible Capacity Reservation (target required).
    # The AwsLaunchTemplate sibling carries only spot/capacity-block -- the
    # interruptible market is instance-level surface at the pinned provider.
    # ForceNew: the purchase option is fixed at launch. Note: the provider
    # keeps capacity-block in state after launch (a re-plan shows no diff --
    # the 6.53.0 perpetual-diff fix).
    market_type = optional(string, "")

    # Request Spot capacity instead of On-Demand -- for interruption-
    # tolerant standalone workloads (a build agent, a batch box). The
    # instance can be reclaimed by AWS with two minutes' notice; pair with
    # instance_interruption_behavior "stop"/"hibernate" (persistent
    # requests) to survive reclaims. Presence implies market_type "spot"
    # when that field is unset. ForceNew: the purchase option is fixed at
    # launch.
    spot_options = optional(object({
      # Maximum price per instance-hour, as a decimal string (e.g. "0.05").
      # AWS default: the On-Demand price -- and leaving it unset is the AWS
      # recommendation, since Spot's discount comes from interruption risk,
      # not bidding.
      max_price = optional(string, "")

      # Spot request type: "one-time" (the AWS default -- the request ends
      # when the instance is interrupted) or "persistent" (EC2 re-requests
      # capacity after interruption; required for "stop"/"hibernate"
      # interruption behavior).
      spot_instance_type = optional(string, "")

      # What happens to the instance on interruption: "terminate" (default),
      # "stop", or "hibernate". "stop"/"hibernate" require a persistent
      # request.
      instance_interruption_behavior = optional(string, "")

      # Expiry of a persistent request, RFC3339 (e.g.
      # "2027-01-01T00:00:00Z"). Only valid when spot_instance_type is
      # "persistent".
      valid_until = optional(string, "")
    }))

    # Target an EC2 Capacity Reservation: "open" (use a matching
    # reservation if one exists -- the AWS default behavior), "none" (never
    # consume a reservation), or a specific reservation / reservation
    # group. Capacity reservations are how latency-critical or
    # failover-critical instances guarantee capacity in an AZ.
    capacity_reservation = optional(object({
      # Reservation preference: "open" (consume a matching reservation when
      # one exists -- AWS's default behavior), "none" (never consume a
      # reservation, even when one matches), or "capacity-reservations-only"
      # (launch ONLY into a matching reservation -- the launch fails when
      # none matches, guaranteeing reserved capacity is what runs). Mutually
      # exclusive with the specific-target fields.
      preference = optional(string, "")

      # Target one specific Capacity Reservation by ID (e.g. "cr-0123...").
      # Mutually exclusive with preference and
      # capacity_reservation_resource_group_arn.
      capacity_reservation_id = optional(string, "")

      # Target a Capacity Reservation resource group by ARN -- a pool of
      # reservations managed together. Mutually exclusive with preference and
      # capacity_reservation_id.
      capacity_reservation_resource_group_arn = optional(string, "")
    }))

    # Where the instance lands: availability zone pinning, placement group
    # membership, tenancy, and Dedicated Host targeting.
    placement = optional(object({
      # Pin the instance to one availability zone (e.g. "us-west-2a"). Unset
      # derives from the subnet (or lets AWS choose in the default VPC).
      availability_zone = optional(string, "")

      # The placement group to launch into, by name: "cluster" groups pack
      # instances for lowest latency, "spread" and "partition" groups
      # separate them for fault isolation. Mutually exclusive with group_id
      # and host_resource_group_arn.
      group_name = optional(string, "")

      # The placement group by ID instead of name. Mutually exclusive with
      # group_name and host_resource_group_arn.
      group_id = optional(string, "")

      # Partition number within a partition placement group.
      partition_number = optional(number, 0)

      # Instance tenancy: "default" (shared hardware), "dedicated"
      # (single-tenant hardware), or "host" (a Dedicated Host -- licensing
      # scenarios).
      tenancy = optional(string, "")

      # Launch onto a specific Dedicated Host by ID (e.g. "h-0123...").
      host_id = optional(string, "")

      # Launch into a host resource group (License Manager-managed Dedicated
      # Hosts) by ARN. Mutually exclusive with placement groups; omit
      # tenancy or set it to "host" alongside this.
      host_resource_group_arn = optional(string, "")
    }))

    # Launch as an AWS Nitro Enclaves parent, enabling isolated enclave VMs
    # for secret-processing workloads. Not supported on every instance
    # type; incompatible with hibernation. ForceNew.
    enclave_enabled = optional(bool, false)

    # Pre-provision the instance for hibernation (requires an encrypted
    # root volume large enough to hold RAM contents). Incompatible with
    # Nitro Enclaves. ForceNew.
    hibernation_enabled = optional(bool, false)

    # Simplified automatic recovery on instance impairment: "default"
    # (recover onto healthy hardware) or "disabled" (leave the instance
    # impaired -- for workloads with instance-store state that recovery
    # would silently discard). Updatable in place.
    auto_recovery = optional(string, "")

    # What an OS-initiated shutdown does to the instance: "stop" (AWS
    # default; the instance can be started again) or "terminate" (the
    # instance is gone). Updatable in place; cannot be set on
    # instance-store-backed instances.
    instance_initiated_shutdown_behavior = optional(string, "")

    # Protect the instance from being stopped via the API -- a guard for
    # long-lived pets. Updatable in place.
    disable_api_stop = optional(bool)

    # Protect the instance from being terminated via the API (termination
    # protection) -- the classic guard against fat-fingered deletion of a
    # stateful pet. Updatable in place; the module's destroy flips it off
    # first only if you remove the protection from the spec.
    disable_api_termination = optional(bool)

    # Allow destroy to proceed even while disable_api_termination /
    # disable_api_stop are true: the engine lifts the protections itself
    # before terminating, instead of failing the destroy. The declarative
    # escape hatch for tearing down a protected pet without first editing
    # its spec. Imported instances always start with this false regardless
    # of the live value (the provider cannot read it back).
    force_destroy = optional(bool, false)

    # Instance user data: a cloud-init config or shell script executed on
    # first boot. Provide PLAIN TEXT here (16 KiB limit before encoding);
    # shell variable syntax like ${HOME} passes through literally. By
    # default, changing user data stops and restarts the instance without
    # replacing it -- and the NEW script does not re-run on an
    # already-initialized instance (cloud-init runs per-instance); set
    # user_data_replace_on_change to get a fresh boot. Mutually exclusive
    # with user_data_base64.
    user_data = optional(string, "")

    # Base64-encoded user data for binary payloads (e.g. gzip-compressed
    # cloud-init) that cannot survive plain-text handling. Mutually
    # exclusive with user_data.
    user_data_base64 = optional(string, "")

    # Replace the instance (destroy and recreate) whenever user data
    # changes, instead of the default stop-update-start. The right choice
    # when user data IS the provisioning mechanism and a stale instance is
    # worse than a new one.
    user_data_replace_on_change = optional(bool, false)
  })
}
