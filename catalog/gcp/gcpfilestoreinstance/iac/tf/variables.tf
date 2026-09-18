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
  description = "GcpFilestoreInstance specification"
  type = object({
    # The GCP project where the Filestore instance is created.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the Filestore instance. This becomes the GCP resource name.
    # Must start with a lowercase letter, contain only lowercase letters, numbers,
    # and hyphens, and be 2-63 characters long. When omitted, metadata.name
    # is used. Immutable after creation.
    instance_name = optional(string, "")

    # Location where the instance will be deployed.
    # For BASIC_HDD, BASIC_SSD, STANDARD, PREMIUM, HIGH_SCALE_SSD, and ZONAL
    # tiers: specify a zone (e.g., "us-central1-a").
    # For ENTERPRISE and REGIONAL tiers: specify a region (e.g., "us-central1").
    # Immutable after creation.
    location = string

    # Service tier controlling performance, availability, and pricing.
    # STANDARD / BASIC_HDD: cost-effective HDD-backed (1 TiB minimum).
    # PREMIUM / BASIC_SSD: mid-tier SSD-backed (2.5 TiB minimum).
    # HIGH_SCALE_SSD: legacy high-performance SSD (10 TiB minimum).
    # ZONAL: modern single-zone SSD with IOPS tuning (1 TiB minimum).
    # REGIONAL: multi-zone SSD with HA (1 TiB minimum).
    # ENTERPRISE: highest tier, regional HA (1 TiB minimum).
    # Immutable after creation.
    tier = string

    # Human-readable description of the instance.
    description = optional(string, "")

    # NFS protocol version.
    # NFS_V3 (default): NFSv3. Broad compatibility, no built-in auth.
    # NFS_V4_1: NFSv4.1. Supports Kerberos security. Available on
    #   HIGH_SCALE_SSD, ZONAL, REGIONAL, and ENTERPRISE tiers.
    # Immutable after creation.
    protocol = optional(string, "")

    # Cloud KMS key for customer-managed encryption at rest (CMEK).
    # Format: projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{key}
    # If not specified, data is encrypted with Google-managed keys.
    # Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # Whether deletion protection is enabled. When true, the instance cannot
    # be deleted until this flag is set to false.
    deletion_protection_enabled = optional(bool, false)

    # Reason for enabling deletion protection. Informational only.
    deletion_protection_reason = optional(string, "")

    # File share configuration. Each Filestore instance has exactly one file share.
    file_share = object({
      # Name of the file share. Becomes the NFS export path.
      # Must start with a letter, followed by letters, numbers, or underscores.
      # Maximum 16 characters.
      # Immutable after creation.
      name = string

      # Capacity of the file share in GiB.
      # Minimum 1024 GiB (1 TiB) for most tiers. BASIC_SSD/PREMIUM requires
      # 2560 GiB minimum. HIGH_SCALE_SSD requires 10240 GiB minimum.
      # The GCP API enforces tier-specific minimums.
      capacity_gb = number

      # NFS export options controlling client access to the file share.
      # Maximum 10 export options per file share.
      # If empty, all clients are allowed with READ_WRITE access and NO_ROOT_SQUASH.
      nfs_export_options = optional(list(object({
        # List of IPv4 addresses or CIDR ranges that are allowed to mount
        # the file share. If empty, all clients are allowed.
        # Maximum 64 IP ranges/addresses across all export options per file share.
        ip_ranges = optional(list(string), [])

        # Access mode for the export.
        # READ_WRITE (default): clients can read and write.
        # READ_ONLY: clients can only read.
        access_mode = optional(string, "")

        # Root squash mode for the export.
        # NO_ROOT_SQUASH (default): root users on clients have root access on the file share.
        # ROOT_SQUASH: root users on clients are mapped to anon_uid/anon_gid.
        squash_mode = optional(string, "")

        # Anonymous user ID used when squash_mode is ROOT_SQUASH.
        # Defaults to 65534 (nobody) if not specified.
        # Only valid when squash_mode is ROOT_SQUASH.
        anon_uid = optional(number)

        # Anonymous group ID used when squash_mode is ROOT_SQUASH.
        # Defaults to 65534 (nogroup) if not specified.
        # Only valid when squash_mode is ROOT_SQUASH.
        anon_gid = optional(number)

        # Source VPC network for ip_ranges, as the network NAME — a
        # GcpVpcNetwork reference resolves to it. Required by GCP for
        # instances using Private Service Connect (where client IPs are not
        # otherwise attributable to a network), optional for other connect
        # modes.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        network = optional(string, "")
      })), [])

      # Restore this file share from an existing Filestore backup, in the
      # format projects/{project}/locations/{location}/backups/{backup}.
      # The share's capacity must be at least the backup's source capacity.
      # Create-time only.
      source_backup = optional(string, "")

      # Restore this file share from a Backup and DR Service backup, in
      # the format projects/{project}/locations/{location}/
      # backupVaults/{vault}/dataSources/{source}/backups/{backup}.
      # The vault-based alternative to source_backup (which restores from
      # Filestore's own backups); set at most one restore source.
      # Create-time only.
      source_backupdr_backup = optional(string, "")
    })

    # VPC network configuration. Each Filestore instance connects to exactly one network.
    network_config = object({
      # VPC network to which the Filestore instance is connected, as the
      # network NAME (e.g. "prod-vpc") — the Filestore API rejects self-link
      # URLs for same-project networks, so the reference resolves the
      # GcpVpcNetwork's plain name output.
      # Immutable after creation.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      network = string

      # Network connection mode.
      # DIRECT_PEERING (default): VPC peering. Simplest setup.
      # PRIVATE_SERVICE_ACCESS: uses a private services connection. Required for
      #   Shared VPC and some enterprise network configurations.
      # PRIVATE_SERVICE_CONNECT: uses Private Service Connect endpoints.
      # Immutable after creation.
      connect_mode = optional(string, "")

      # A /29 CIDR block for internal IP addresses reserved for this instance.
      # Must be unique and non-overlapping with existing subnets in the VPC.
      # If not specified, GCP automatically selects an unused range.
      # Immutable after creation.
      reserved_ip_range = optional(string, "")

      # IP address versions the instance serves. Values: "MODE_IPV4",
      # "MODE_IPV6". When empty, ["MODE_IPV4"] is used — the standard NFS
      # posture. Immutable after creation.
      modes = optional(list(string), [])

      # Consumer project in which the Private Service Connect endpoint is
      # created — a project ID; a GcpProject reference resolves to it. If
      # omitted, the endpoint is created in the instance's own project.
      # Only meaningful when connect_mode is PRIVATE_SERVICE_CONNECT
      # (enforced pre-deploy). Immutable after creation.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      psc_endpoint_project = optional(string, "")
    })

    # Performance configuration for IOPS tuning.
    # Available on ZONAL, REGIONAL, and ENTERPRISE tiers.
    # If not specified, the instance uses the default performance for its tier.
    performance_config = optional(object({
      # Fixed IOPS provisioning. IOPS remains constant regardless of capacity.
      # Mutually exclusive with iops_per_tb.
      fixed_iops = optional(object({
        # The number of IOPS to provision. Must be a multiple of 1000.
        max_iops = number
      }))

      # Dynamic IOPS provisioning. IOPS scales with instance capacity.
      # Mutually exclusive with fixed_iops.
      iops_per_tb = optional(object({
        # Maximum IOPS per terabyte of capacity.
        max_iops_per_tb = number
      }))
    }))

    # Cross-instance replication established at create time: this instance
    # becomes the STANDBY replica of an existing ACTIVE peer (the common
    # DR posture). Create-time only.
    initial_replication = optional(object({
      # Replication role of THIS instance:
      #   ""        -- same as "STANDBY" (GCP default; this instance receives
      #                replication from the peer)
      #   "STANDBY" -- this instance is the read-only replica
      #   "ACTIVE"  -- this instance is the replication source
      role = optional(string, "")

      # Peer Filestore instances in the replication relationship, each as a
      # reference to a GcpFilestoreInstance (or a literal full resource path
      # projects/{project}/locations/{location}/instances/{instance}).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      peer_instances = list(string)
    }))

    # User labels merged with Planton attribution labels (which win on key
    # conflicts). Keys and values must match GCP label constraints.
    labels = optional(map(string), {})

    # Resource Manager tags bound to the instance for org-policy and IAM
    # conditions. Keys in the form "tagKeys/{id}", values "tagValues/{id}".
    # Create-time only.
    tags = optional(map(string), {})

    # LDAP directory integration for NFSv4.1 identity mapping. Requires
    # protocol NFS_V4_1 — with NFSv3, identity is numeric UID/GID
    # matching and no directory service applies.
    ldap = optional(object({
      # LDAP domain name, e.g. "my-domain.com".
      domain = string

      # LDAP server addresses — either all DNS names (e.g.
      # "ldap.example.com") or all IP addresses; GCP rejects a mix of the
      # two formats.
      servers = list(string)

      # Groups Organizational Unit (OU) — an optional hint that narrows
      # LDAP lookups to one OU instead of querying the whole namespace
      # (faster lookups on large directories).
      groups_ou = optional(string, "")

      # Users Organizational Unit (OU) — the same lookup-narrowing hint
      # for user entries.
      users_ou = optional(string, "")
    }))

    # Desired state of THIS instance's replica relationship, when the
    # instance is the STANDBY side of a replication pair:
    #   "READY"  (default) -- replication runs; the standby receives
    #                         changes from the active peer
    #   "PAUSED"           -- replication is paused (e.g. to freeze the
    #                         standby at a point in time); resume by
    #                         setting READY again
    # Updatable in place; has no effect on an instance without a
    # replica relationship.
    desired_replica_state = optional(string)

    # Deletion policy for the instance — what happens when this resource
    # is destroyed (evaluated only after deletion_protection_enabled
    # allows the destroy at all):
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the instance and every file on its share are deleted
    #   "PREVENT" -- destroy FAILS; a second, independent guard for a
    #                file server whose data exists nowhere else
    #   "ABANDON" -- the instance is removed from management but left
    #                running (and billing) in GCP with its data intact
    deletion_policy = optional(string, "")
  })
}
