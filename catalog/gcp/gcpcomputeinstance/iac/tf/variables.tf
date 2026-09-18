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
  description = "GcpComputeInstance specification"
  type = object({
    # The GCP project that owns the instance.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the Compute Engine instance. 1-63 characters, lowercase
    # letters, numbers, and hyphens; must start with a letter and cannot end
    # with a hyphen. When omitted, metadata.name is used.
    # Immutable after creation.
    instance_name = optional(string, "")

    # Zone where the instance runs, e.g. "us-central1-a".
    # Immutable after creation (moving zones replaces the VM).
    zone = string

    # Machine type, e.g. "e2-medium", "n2-standard-4", "c3-highcpu-8", or a
    # custom shape like "custom-6-20480". Mutable — changing it stops and
    # restarts the VM (requires allow_stopping_for_update).
    machine_type = string

    # Human-readable description of the instance.
    description = optional(string, "")

    # Custom fully-qualified DNS hostname, e.g. "db-1.prod.internal".
    # RFC-1035 labels separated by dots; when unset GCP derives
    # "<name>.c.<project>.internal". Immutable after creation.
    hostname = optional(string, "")

    # Boot disk configuration — the disk the OS boots from.
    boot_disk = object({
      # Source image for a fresh boot disk. Accepts an image family
      # ("debian-cloud/debian-12", "ubuntu-os-cloud/ubuntu-2404-lts-amd64") or
      # a specific image self link. Families resolve to the newest image at
      # create time. Create-time only.
      image = optional(string, "")

      # Source snapshot to restore the boot disk from (name or self link).
      # Create-time only.
      source_snapshot = optional(string, "")

      # Existing bootable disk to boot from, referenced as a GcpComputeDisk.
      # The disk must live in the instance's zone. When booting from an
      # existing disk, size/type/encryption below are ignored — the disk
      # already owns them.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      source_disk = optional(string, "")

      # Size of the boot disk in GB. When omitted, the image or snapshot size
      # is used. Grows in place; never shrinks. Most OS images need at least
      # 10 GB; the API floor itself is 1 GB.
      size_gb = optional(number, 0)

      # Disk type: "pd-standard" (HDD), "pd-balanced" (default in GCP for most
      # machine shapes and the sensible choice), "pd-ssd" (high IOPS), or a
      # hyperdisk type on supported machine families ("hyperdisk-balanced").
      type = optional(string, "")

      # Delete the boot disk automatically when the instance is deleted.
      # Defaults to true (matching GCP). Set false to keep the OS disk for
      # forensics or re-attachment.
      auto_delete = optional(bool)

      # Device name exposed under /dev/disk/by-id/google-*. When omitted GCP
      # assigns one.
      device_name = optional(string, "")

      # Customer-managed encryption key (CMEK) for the boot disk, referenced
      # as a GcpKmsKey. The Compute Engine service agent
      # (service-<project-number>@compute-system.iam.gserviceaccount.com) must
      # hold roles/cloudkms.cryptoKeyEncrypterDecrypter on the key.
      # Create-time only.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key = optional(string, "")

      # Labels applied to the boot disk itself (distinct from instance
      # labels) — useful for disk-level cost attribution and snapshot
      # policies.
      disk_labels = optional(map(string), {})

      # Provisioned IOPS for hyperdisk types that support tuning
      # (e.g. hyperdisk-extreme, hyperdisk-balanced). Leave unset for pd-*
      # types.
      provisioned_iops = optional(number)

      # Provisioned throughput in MB/s for hyperdisk types that support
      # tuning (e.g. hyperdisk-throughput, hyperdisk-balanced). Leave unset
      # for pd-* types.
      provisioned_throughput = optional(number)

      # CPU architecture of the disk/image: "X86_64" or "ARM64" (e.g. for
      # Tau T2A/Axion machine types). Normally inferred from the image.
      architecture = optional(string, "")

      # Create the boot disk in confidential-compute mode (hyperdisk SKUs
      # only; requires kms_key).
      enable_confidential_compute = optional(bool, false)

      # Self links of resource policies to attach to the boot disk at create
      # time (e.g. a snapshot schedule). GCP currently allows at most one.
      # Changing this replaces the instance.
      resource_policies = optional(list(string), [])

      # URL of the storage pool to create the boot disk in (hyperdisk storage
      # pools).
      storage_pool = optional(string, "")

      # Attachment mode: "READ_WRITE" (default) or "READ_ONLY" (share one
      # boot disk read-only across many VMs).
      mode = optional(string, "")

      # Disk attachment interface: "SCSI" or "NVME". GCP normally selects
      # the right interface from the machine type and disk type — the
      # provider's own guidance is "only used for specific cases, please
      # don't specify this field without advice from Google". Leave unset
      # unless you have such a case.
      interface = optional(string, "")

      # Force-attach a REGIONAL boot disk even if it is currently attached
      # to another instance (regional-disk failover takeover). Attempting to
      # force-attach a zonal disk fails. Changing this replaces the VM.
      force_attach = optional(bool, false)

      # Guest OS features to enable on the boot disk, e.g.
      # ["UEFI_COMPATIBLE", "SECURE_BOOT", "GVNIC", "MULTI_IP_SUBNET",
      # "WINDOWS"]. The accepted set evolves with GCP — see "Enabling guest
      # operating system features" in the Compute Engine docs. Create-time
      # only. When set, list the image's COMPLETE feature set, never just
      # the additions: the API merges this list with the image's own
      # features at create and the stored disk echoes the merged set, which
      # the provider then compares authoritatively with replace-on-change
      # semantics — a partial list plans a VM REPLACEMENT on every re-apply
      # (live-verified: debian-12's ["UEFI_COMPATIBLE", "GVNIC"] echoed back
      # ["UEFI_COMPATIBLE", "VIRTIO_SCSI_MULTIQUEUE", "GVNIC", "SEV_CAPABLE",
      # "SEV_LIVE_MIGRATABLE_V2"]). Leave unset to follow the image's own
      # features cleanly.
      guest_os_features = optional(list(string), [])

      # Zones for a REGIONAL boot disk (exactly two, one of which must be
      # the instance's own zone; short names or self links). Setting this
      # converts the boot disk to a regional disk replicated across both
      # zones. Only valid with a source_snapshot boot source (enforced
      # pre-deploy): GCP rejects creating a regional boot disk from an
      # image (live-verified API 400: "Creating a regional disk from a
      # source image is not supported yet" — the snapshot path was
      # live-verified to produce a true regional boot disk), and a
      # pre-created source_disk already carries its own zones.
      # Create-time only.
      replica_zones = optional(list(string), [])

      # Resource Manager tags bound to the boot disk at create time. Keys in
      # the form "tagKeys/{id}", values "tagValues/{id}". Create-time only —
      # changing them replaces the VM. Ignored when booting from an existing
      # source_disk.
      resource_manager_tags = optional(map(string), {})

      # Service account used for the encryption request of kms_key (CMEK).
      # When omitted, the Compute Engine default service agent is used.
      # Only meaningful together with kms_key.
      kms_key_service_account = optional(string, "")

      # Decrypts the source image when it is itself CMEK-encrypted. Only
      # valid together with image.
      source_image_encryption = optional(object({
        # The KMS key the source was encrypted with, referenced as a GcpKmsKey
        # or a literal self link. The service agent performing the read needs
        # roles/cloudkms.cryptoKeyEncrypterDecrypter on it.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        kms_key = string

        # Service account used for the decryption request. When omitted, the
        # Compute Engine default service agent is used.
        kms_key_service_account = optional(string, "")
      }))

      # Decrypts the source snapshot when it is itself CMEK-encrypted. Only
      # valid together with source_snapshot.
      source_snapshot_encryption = optional(object({
        # The KMS key the source was encrypted with, referenced as a GcpKmsKey
        # or a literal self link. The service agent performing the read needs
        # roles/cloudkms.cryptoKeyEncrypterDecrypter on it.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        kms_key = string

        # Service account used for the decryption request. When omitted, the
        # Compute Engine default service agent is used.
        kms_key_service_account = optional(string, "")
      }))
    })

    # Additional persistent data disks attached to the instance. Each disk
    # is a first-class GcpComputeDisk resource referenced by self link — the
    # disk has its own lifecycle and survives this VM unless its own
    # configuration says otherwise.
    attached_disks = optional(list(object({
      # The disk to attach, referenced as a GcpComputeDisk (or a literal disk
      # name/self link). The disk must live in the instance's zone.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      source = string

      # Device name exposed under /dev/disk/by-id/google-*. When omitted GCP
      # assigns "persistent-disk-N".
      device_name = optional(string, "")

      # Attachment mode: "READ_WRITE" (default) or "READ_ONLY" (lets many VMs
      # attach the same disk simultaneously).
      mode = optional(string, "")

      # The CMEK key protecting the attached disk, referenced as a GcpKmsKey.
      # Required only when the disk is CMEK-encrypted — the attachment must
      # present the same key the disk was created with.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key = optional(string, "")

      # Service account used for the encryption request of kms_key (CMEK).
      # When omitted, the Compute Engine default service agent is used.
      # Only meaningful together with kms_key.
      kms_key_service_account = optional(string, "")

      # Force-attach a REGIONAL disk even if it is currently attached to
      # another instance (regional-disk failover takeover). Attempting to
      # force-attach a zonal disk fails. Changing this replaces the VM.
      force_attach = optional(bool, false)
    })), [])

    # Ephemeral local-SSD scratch disks physically attached to the host.
    # Contents are lost when the VM stops or is preempted — use only for
    # caches, temp space, and high-IOPS scratch data. Create-time only.
    scratch_disks = optional(list(object({
      # Disk interface: "NVME" (recommended; highest performance) or "SCSI".
      interface = string

      # Size in GB. Local SSDs come in fixed 375 GB units (or 3000 GB on
      # supported Z3 shapes). When omitted, 375 is used.
      size_gb = optional(number, 0)

      # Device name exposed under /dev/disk/by-id/google-*.
      device_name = optional(string, "")
    })), [])

    # Network interfaces. At least one is required; multiple NICs must each
    # attach to a different VPC network. NIC count and their target
    # networks are immutable after creation.
    network_interfaces = list(object({
      # VPC network for this interface, referenced as a GcpVpcNetwork.
      # Sufficient alone only for auto-mode VPCs; custom-mode VPCs need
      # subnetwork.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      network = optional(string, "")

      # Subnetwork for this interface, referenced as a GcpSubnetwork. The
      # subnetwork's region must contain the instance's zone.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnetwork = optional(string, "")

      # Project owning the subnetwork — set when attaching to a Shared VPC
      # host project's subnetwork from a service project.
      subnetwork_project = optional(string, "")

      # Static internal IP for this interface. Accepts a literal IP or a
      # reference to a reserved INTERNAL GcpAddress. When omitted, GCP
      # assigns an ephemeral internal IP from the subnetwork range.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      network_ip = optional(string, "")

      # External IPv4 access configs. Empty means no external IP (private
      # VM — pair with Cloud NAT for egress). GCP supports at most one
      # access config per interface.
      access_configs = optional(list(object({
        # Static external IP, as a literal or a reference to a reserved
        # EXTERNAL GcpAddress. Alternative to ephemeral: true.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        nat_ip = optional(string, "")

        # Network service tier for this IP: "PREMIUM" (default; Google's global
        # backbone) or "STANDARD" (regional, cheaper).
        network_tier = optional(string, "")

        # Domain name for the public PTR (reverse DNS) record of this IP.
        public_ptr_domain_name = optional(string, "")
      })), [])

      # External IPv6 access configs. Requires stack_type "IPV4_IPV6" and a
      # subnetwork with an external IPv6 range. At most one per interface.
      ipv6_access_configs = optional(list(object({
        # Network service tier for IPv6 traffic. Only "PREMIUM" is valid.
        network_tier = string

        # Domain name for the public PTR (reverse DNS) record of the external
        # IPv6 range.
        public_ptr_domain_name = optional(string, "")

        # Static EXTERNAL IPv6 address (the first address of the external
        # range) to pin to this interface. Must be unused and in the same
        # region as the instance's zone. When omitted, GCP assigns an external
        # IPv6 range from the subnetwork. Changing it replaces the VM.
        external_ipv6 = optional(string, "")

        # Prefix length of the external IPv6 range (the provider models this
        # as a string). Normally read back from GCP rather than set — only
        # meaningful together with external_ipv6. Changing it replaces the VM.
        external_ipv6_prefix_length = optional(string, "")

        # Name of this IPv6 access configuration; GCP's recommended value is
        # "External IPv6". When omitted, GCP assigns one. Changing it replaces
        # the VM.
        name = optional(string, "")
      })), [])

      # IP stack of the interface:
      #   ""            -- same as "IPV4_ONLY" (GCP default)
      #   "IPV4_ONLY"   -- IPv4 addresses only
      #   "IPV4_IPV6"   -- dual stack (subnetwork must have an IPv6 range)
      #   "IPV6_ONLY"   -- IPv6 only (supported on IPv6-enabled subnetworks)
      stack_type = optional(string, "")

      # vNIC type: "" (GCP picks), "GVNIC" (recommended on modern machine
      # families; required for TIER_1 bandwidth), "VIRTIO_NET" (legacy), or
      # the RDMA types "IDPF", "MRDMA", "IRDMA" on specialized shapes.
      nic_type = optional(string, "")

      # Networking queue count for Rx and Tx (1-32). When omitted GCP sizes
      # queues from vCPU count.
      queue_count = optional(number)

      # Alias IP ranges served by this interface — the mechanism behind
      # per-pod/per-container IPs and multi-IP VMs.
      alias_ip_ranges = optional(list(object({
        # The alias range: a CIDR ("10.1.2.0/24"), a single IP ("10.1.2.3"), or
        # a netmask ("/24") to auto-allocate from the range.
        ip_cidr_range = string

        # Secondary range name on the subnetwork to allocate from. When omitted
        # the primary range is used.
        subnetwork_range_name = optional(string, "")
      })), [])

      # URL of a Private Service Connect NETWORK ATTACHMENT this interface
      # connects to, in the form
      # "projects/{projectNumber}/regions/{region}/networkAttachments/{name}"
      # — the consumer side of PSC interfaces, connecting this VM into a
      # producer's VPC. An attachment-only interface is legal (no network or
      # subnetwork). Immutable after creation.
      network_attachment = optional(string, "")

      # VLAN tag (2-255) making this a DYNAMIC network interface — a
      # sub-interface multiplexed onto a parent NIC. Immutable after
      # creation.
      vlan = optional(number)

      # IGMP multicast query support on this interface:
      #   ""                    -- GCP default (disabled)
      #   "IGMP_QUERY_V2"       -- IGMPv2 queries enabled (multicast)
      #   "IGMP_QUERY_DISABLED" -- explicitly disabled
      # Updatable in place.
      igmp_query = optional(string, "")

      # Static INTERNAL IPv6 address for this interface (requires an
      # IPv6-enabled stack_type and subnetwork). When omitted, GCP assigns
      # one from the subnetwork's internal IPv6 range. Long-form and
      # compressed spellings are equivalent.
      ipv6_address = optional(string, "")

      # Prefix length of the primary internal IPv6 range assigned to this
      # interface. When omitted, GCP assigns its default.
      internal_ipv6_prefix_length = optional(number)
    }))

    # Service account the VM's workloads authenticate as. When omitted, the
    # Compute Engine default service account is used with its default
    # scopes — prefer a dedicated least-privilege account for production.
    service_account = optional(object({
      # Service account email, referenced as a GcpServiceAccount or a literal
      # email. Changing it stops and restarts the VM (requires
      # allow_stopping_for_update).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      email = optional(string, "")

      # OAuth scopes for the attached account. The modern practice is a
      # single "https://www.googleapis.com/auth/cloud-platform" scope with
      # access controlled entirely by IAM roles; narrower legacy scopes
      # remain supported. Required when this block is set.
      scopes = list(string)
    }))

    # Scheduling policy: Spot vs standard provisioning, maintenance
    # behavior, run-duration limits, and sole-tenant node placement.
    scheduling = optional(object({
      # Provisioning model:
      #   ""                  -- same as "STANDARD"
      #   "STANDARD"          -- on-demand capacity
      #   "SPOT"              -- deeply discounted preemptible capacity; GCP
      #                          may reclaim the VM at any time
      #   "FLEX_START"        -- discounted capacity with a flexible start
      #                          time (Dynamic Workload Scheduler); requires
      #                          max_run_duration_seconds and is deleted when
      #                          reclaimed
      #   "RESERVATION_BOUND" -- runs only on capacity from one specific
      #                          reservation (pair with reservation_affinity
      #                          type SPECIFIC_RESERVATION)
      # Create-time only.
      provisioning_model = optional(string, "")

      # Restart the VM automatically when Compute Engine (not a user) stops
      # it. Defaults to true for standard VMs; must be false (or unset) for
      # Spot.
      automatic_restart = optional(bool)

      # Host maintenance behavior: "" (GCP default "MIGRATE"), "MIGRATE"
      # (live-migrate; zero downtime), or "TERMINATE" (stop during
      # maintenance — required for GPUs and confidential VMs).
      on_host_maintenance = optional(string, "")

      # What GCP does when the VM is reclaimed (Spot preemption, FLEX_START
      # expiry) or a run-duration limit fires: "STOP" (keep the stopped VM
      # and disks) or "DELETE" (remove the VM). Applies to SPOT and
      # FLEX_START models and to timed-run VMs
      # (max_run_duration_seconds/termination_time).
      instance_termination_action = optional(string, "")

      # Maximum run duration in seconds, after which
      # instance_termination_action is executed. The duration clock starts at
      # every VM start. Mutually exclusive with termination_time.
      max_run_duration_seconds = optional(number)

      # Absolute timestamp (RFC 3339) at which the VM is terminated. Mutually
      # exclusive with max_run_duration_seconds.
      termination_time = optional(string, "")

      # Discard local-SSD contents when the VM is stopped by a lifetime limit
      # (max_run_duration/termination_time) instead of preserving them.
      discard_local_ssds_on_stop = optional(bool)

      # Availability domain for spread-placement within the zone (used with
      # spread placement policies). 0 means unset.
      availability_domain = optional(number)

      # Minimum vCPUs on the sole-tenant node this VM can be scheduled onto.
      # Sole-tenancy only.
      min_node_cpus = optional(number)

      # Sole-tenant node affinities selecting which node groups this VM may
      # run on. Setting any affinity places the VM on sole-tenant hardware.
      node_affinities = optional(list(object({
        # Node-group label key to match (e.g.
        # "compute.googleapis.com/node-group-name").
        key = string

        # Match operator: "IN" or "NOT_IN".
        operator = string

        # Label values to match.
        values = list(string)
      })), [])

      # How long Compute Engine waits for a local-SSD-preserving recovery
      # when the host fails, in seconds, before falling back to default
      # recovery.
      local_ssd_recovery_timeout_seconds = optional(number)

      # How long Compute Engine waits, in seconds, before declaring the host
      # failed and starting host-error recovery (restart or termination per
      # automatic_restart). A lower value recovers faster from a hung host at
      # the cost of more false positives; leave unset for Compute Engine's
      # default recovery timing. Must be 90..330 in steps of 30 (90, 120, ...,
      # 330).
      host_error_timeout_seconds = optional(number)
    }))

    # Shielded VM configuration (secure boot, vTPM, integrity monitoring).
    # Requires an image with Shielded VM support (all recent Google-provided
    # images qualify).
    shielded_instance_config = optional(object({
      # Verify the boot loader's signature chain; blocks boot on tampering.
      # GCP default is false because some third-party images are unsigned.
      enable_secure_boot = optional(bool)

      # Virtual Trusted Platform Module. GCP default is true.
      enable_vtpm = optional(bool)

      # Boot-integrity measurement and monitoring via the vTPM. GCP default
      # is true.
      enable_integrity_monitoring = optional(bool)
    }))

    # Confidential VM configuration — hardware memory encryption (AMD SEV /
    # SEV-SNP or Intel TDX). Requires a supported machine family (e.g. N2D,
    # C2D, C3) and on_host_maintenance = "TERMINATE". Create-time only.
    confidential_instance_config = optional(object({
      # Confidential computing technology:
      #   "SEV"     -- AMD Secure Encrypted Virtualization (N2D/C2D/C3D)
      #   "SEV_SNP" -- AMD SEV Secure Nested Paging (requires an AMD Milan+
      #                min_cpu_platform)
      #   "TDX"     -- Intel Trust Domain Extensions (C3)
      confidential_instance_type = string
    }))

    # Advanced machine features: nested virtualization, SMT control, visible
    # core count, UEFI networking, performance monitoring unit, and turbo
    # mode.
    advanced_machine_features = optional(object({
      # Expose nested virtualization support (VMX) to the guest — run VMs
      # inside this VM.
      enable_nested_virtualization = optional(bool)

      # Threads per physical core: 1 disables simultaneous multithreading
      # (SMT) — common for licensing and security isolation; 2 is the
      # hardware default.
      threads_per_core = optional(number)

      # Number of physical cores exposed to the guest (core visibility for
      # per-core licensing). When unset all cores are visible.
      visible_core_count = optional(number)

      # Enable UEFI networking in the guest firmware. Create-time only.
      enable_uefi_networking = optional(bool)

      # Performance monitoring unit exposure level: "STANDARD", "ENHANCED",
      # or "ARCHITECTURAL".
      performance_monitoring_unit = optional(string, "")

      # Turbo frequency mode. "ALL_CORE_MAX" runs all cores at maximum turbo
      # frequency (supported machine families only).
      turbo_mode = optional(string, "")
    }))

    # GPU accelerator cards attached to the instance. Requires a
    # GPU-capable zone and on_host_maintenance = "TERMINATE". Each entry
    # attaches at least one card (count >= 1). Removing every entry from an
    # existing VM leaves its GPUs attached (the provider preserves the
    # current accelerators when the block is absent); detaching GPUs
    # replaces the VM, so plan it as a recreate rather than an edit.
    guest_accelerators = optional(list(object({
      # Accelerator type available in the instance's zone, e.g.
      # "nvidia-tesla-t4", "nvidia-l4", "nvidia-a100-80gb".
      type = string

      # Number of cards of this type.
      count = number
    })), [])

    # Reservation affinity — whether this VM consumes capacity from any
    # matching reservation, a specific reservation, or none.
    reservation_affinity = optional(object({
      # Reservation consumption mode:
      #   "ANY_RESERVATION"      -- consume any matching reservation (GCP
      #                             default)
      #   "SPECIFIC_RESERVATION" -- consume only the named reservation
      #   "NO_RESERVATION"       -- never consume reserved capacity
      type = string

      # The specific reservation to consume (type SPECIFIC_RESERVATION).
      specific_reservation = optional(object({
        # Reservation label key — use
        # "compute.googleapis.com/reservation-name" to target by name.
        key = string

        # Reservation label values (the reservation name when using the
        # reservation-name key).
        values = list(string)
      }))
    }))

    # Per-VM egress bandwidth tier. "TIER_1" raises the bandwidth cap on
    # supported machine shapes (N2/N2D/C2/C3 with >= 30 vCPUs and gVNIC);
    # "DEFAULT" is the standard cap.
    total_egress_bandwidth_tier = optional(string, "")

    # Custom metadata key/value pairs made available to the guest OS via the
    # metadata server. Well-known keys configure agents and features (e.g.
    # "enable-oslogin", "startup-script-url").
    metadata = optional(map(string), {})

    # Startup script executed by the guest agent on every boot. Maps to the
    # metadata_startup_script surface, which keeps it distinct from user
    # metadata and re-runs it on each start.
    startup_script = optional(string, "")

    # SSH public keys in "username:ssh-rsa AAAA... user" format. Folded into
    # the instance metadata "ssh-keys" key (newline-joined) identically on
    # both engines. Ignored by VMs using OS Login.
    ssh_keys = optional(list(string), [])

    # User labels merged with Planton attribution labels (which win on key
    # conflicts). Keys and values must match GCP label constraints:
    # lowercase letters, numbers, hyphens, underscores; keys start with a
    # letter, max 63 characters.
    labels = optional(map(string), {})

    # Network tags used by firewall rules and network routes to select this
    # instance.
    tags = optional(list(string), [])

    # Resource Manager tags bound to the instance for org-policy and IAM
    # conditions. Keys in the form "tagKeys/{id}", values "tagValues/{id}".
    # Create-time only.
    resource_manager_tags = optional(map(string), {})

    # Self links of compute resource policies to attach to the instance
    # (e.g. an instance schedule that starts/stops the VM on a calendar).
    # GCP currently allows at most one policy per instance.
    resource_policies = optional(list(string), [])

    # Minimum CPU platform for the VM, e.g. "Intel Ice Lake" or
    # "AMD Milan". Constrains scheduling to hosts with at least this
    # platform.
    min_cpu_platform = optional(string, "")

    # Allow sending/receiving packets with source or destination IPs that do
    # not match the instance's own addresses — required for VMs acting as
    # routers, NAT gateways, or VPN endpoints. Create-time only.
    can_ip_forward = optional(bool, false)

    # Enable the virtual display device (needed by some remote-desktop and
    # screen-capture tooling on headless VMs).
    enable_display = optional(bool, false)

    # Protect the instance against accidental deletion. Deleting a protected
    # instance fails until this is set back to false. Defaults to false: a
    # VM is a compute node whose data levers are its disks — the boot disk's
    # auto_delete and each GcpComputeDisk's own lifecycle guard the data.
    deletion_protection = optional(bool, false)

    # Desired lifecycle status of the VM:
    #   ""           -- same as "RUNNING"
    #   "RUNNING"    -- started
    #   "SUSPENDED"  -- suspended to disk (fast resume; memory persisted)
    #   "TERMINATED" -- stopped (compute billing stops; disks keep billing)
    # Changing this starts/suspends/stops the VM in place.
    desired_status = optional(string, "")

    # Allow the provider to stop and restart the VM when an update requires
    # it (machine type, service account, network interface changes, ...).
    # Without it those updates fail instead of causing downtime. Recommended
    # true for instances whose brief restart is acceptable.
    allow_stopping_for_update = optional(bool)

    # Action GCP takes on the VM when a Cloud KMS key protecting it is
    # revoked: "NONE" (default) or "STOP".
    key_revocation_action_type = optional(string, "")

    # Customer-managed encryption key (CMEK) for INSTANCE-LEVEL data —
    # memory contents and other instance state, distinct from the per-disk
    # keys on boot_disk/attached_disks. Create-time only.
    instance_encryption_key = optional(object({
      # The KMS key encrypting instance-level data, referenced as a
      # GcpKmsKey or a literal self link. The Compute Engine service agent
      # needs roles/cloudkms.cryptoKeyEncrypterDecrypter on it.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key = string

      # Service account used for the encryption request. When omitted, the
      # Compute Engine default service agent is used.
      kms_key_service_account = optional(string, "")
    }))

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the instance is deleted (disks follow their own
    #                lifecycle: boot auto_delete, GcpComputeDisk configs)
    #   "PREVENT" -- destroy FAILS; a guard rail beyond deletion_protection
    #                because it blocks the IaC destroy itself
    #   "ABANDON" -- the instance is removed from management but left
    #                running in GCP
    deletion_policy = optional(string, "")

    # Managed workload identity for the VM: a SPIFFE identity issued to the
    # instance (and, optionally, X.509 identity certificates) so workloads
    # on it authenticate to each other by identity instead of shared
    # secrets or network position. Create-time only: both fields are
    # immutable, so changing them replaces the VM.
    workload_identity_config = optional(object({
      # The SPIFFE ID Compute Engine issues to the instance, e.g.
      # "spiffe://PROJECT.svc.id.goog/ns/NAMESPACE/sa/SERVICE_ACCOUNT" or a
      # workload-identity-pool identity of the form
      # "spiffe://POOL.global.PROJECT_NUMBER.workload.id.goog/ns/NS/sa/SA".
      # Immutable.
      identity = string

      # Whether Compute Engine also issues and rotates X.509 certificates
      # bound to the identity, made available on the VM for mutual TLS.
      # Immutable.
      identity_certificate_enabled = optional(bool, false)
    }))
  })
}
