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
  description = "GcpComputeMig specification"
  type = object({
    # The GCP project that owns every resource in the group.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the managed instance group. 1-63 characters, lowercase
    # letters, numbers, and hyphens; must start with a letter and cannot
    # end with a hyphen. When omitted, metadata.name is used. The name
    # also seeds the managed resources around the group (template name
    # prefix, autoscaler name). Immutable after creation.
    mig_name = optional(string, "")

    # Zone for a ZONAL group (e.g. "us-central1-a") — every VM runs in
    # this one zone. Exactly one of zone or region must be set.
    # Immutable: a group cannot move between scopes or locations.
    zone = optional(string, "")

    # Region for a REGIONAL group (e.g. "us-central1") — VMs are spread
    # across the region's zones (see distribution_policy) so a zone
    # outage takes down only part of the fleet. Exactly one of zone or
    # region must be set. Immutable: a group cannot move between scopes
    # or locations.
    region = optional(string, "")

    # Human-readable description of the group shown in the console.
    # Immutable after creation.
    description = optional(string, "")

    # Base name for VMs created by the group — instances are named
    # "<base_instance_name>-<random suffix>" (e.g. "web-x7kq"). RFC-1035
    # format, 1-58 characters. When omitted, the group name (mig_name or
    # metadata.name) is used.
    base_instance_name = optional(string, "")

    # The instance template — what every VM in the group looks like:
    # machine type, disks, networking, identity, and scheduling. The
    # template is IMMUTABLE in GCP (only labels can change in place):
    # every other change here creates a NEW template and rolls the group
    # to it per update_policy. See GcpComputeMigTemplate for the
    # replace-on-change semantics of each field.
    template = object({
      # Machine type for every VM, e.g. "e2-micro", "e2-medium",
      # "n2-standard-4", or a custom shape like "custom-6-20480".
      # Changing it rotates the template.
      machine_type = string

      # Human-readable description of the TEMPLATE resource itself.
      # Changing it rotates the template.
      description = optional(string, "")

      # Description stamped onto each INSTANCE created from the template
      # (visible on the VMs, distinct from the template's own
      # description). Changing it rotates the template.
      instance_description = optional(string, "")

      # Disks attached to every VM created from the template. Exactly one
      # disk must be the boot disk. Each entry either creates a new disk
      # per instance (from an image, a snapshot, or blank) or attaches one
      # existing disk (source) — see GcpComputeMigTemplateDisk.
      # Changing disks rotates the template.
      disks = list(object({
        # Whether this is the boot disk — the disk the VMs boot from.
        # Exactly one disk in the template must set this.
        boot = optional(bool, false)

        # Source image for a fresh disk on each VM. Accepts an image family
        # ("debian-cloud/debian-12", "ubuntu-os-cloud/ubuntu-2404-lts-amd64")
        # or a specific image self link. Families resolve to the newest image
        # at template creation.
        source_image = optional(string, "")

        # Source snapshot each VM's disk is restored from (name or self
        # link).
        source_snapshot = optional(string, "")

        # Existing disk to attach (all VMs share it), referenced as a
        # GcpComputeDisk or a literal disk name/self link. Shared attachment
        # requires mode READ_ONLY.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        source = optional(string, "")

        # Size of the disk in GB. When omitted, the image or snapshot size is
        # used (blank disks require a size).
        size_gb = optional(number, 0)

        # Disk type: "pd-standard" (HDD), "pd-balanced" (default in GCP and
        # the sensible choice), "pd-ssd" (high IOPS), local "local-ssd"
        # (ephemeral scratch), or a hyperdisk type on supported machine
        # families ("hyperdisk-balanced").
        disk_type = optional(string, "")

        # Disk role: "PERSISTENT" (default) or "SCRATCH" (ephemeral local
        # SSD; contents lost when the VM stops — pair with disk_type
        # "local-ssd" and 375 GB units).
        type = optional(string, "")

        # Delete each VM's disk automatically when the VM is deleted.
        # Defaults to true (matching GCP). Stateful disks (see the spec's
        # stateful_disks) override this per device.
        auto_delete = optional(bool)

        # Device name exposed under /dev/disk/by-id/google-*. When omitted
        # GCP assigns one. Stateful disk rules match on this name.
        device_name = optional(string, "")

        # Name for disks created per VM. When omitted GCP derives one from
        # the instance name. Rarely needed.
        disk_name = optional(string, "")

        # Attachment mode: "READ_WRITE" (default) or "READ_ONLY" (required
        # when attaching one existing source disk across the fleet).
        mode = optional(string, "")

        # Disk attachment interface: "SCSI" or "NVME". GCP normally selects
        # the right interface from the machine type and disk type — the
        # provider's own guidance is to leave it unset without advice from
        # Google.
        interface = optional(string, "")

        # Labels applied to the per-VM disks (distinct from VM labels) —
        # disk-level cost attribution and snapshot policies.
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

        # Guest OS features to enable on the disk, e.g. ["UEFI_COMPATIBLE",
        # "SECURE_BOOT", "GVNIC", "MULTI_IP_SUBNET", "WINDOWS"]. The accepted
        # set evolves with GCP — see "Enabling guest operating system
        # features" in the Compute Engine docs.
        guest_os_features = optional(list(string), [])

        # Self links of resource policies attached to each per-VM disk (e.g.
        # a snapshot schedule). GCP currently allows at most one per disk.
        resource_policies = optional(list(string), [])

        # Resource Manager tags bound to the per-VM disks. Keys in the form
        # "tagKeys/{id}", values "tagValues/{id}".
        resource_manager_tags = optional(map(string), {})

        # URL of the storage pool to create per-VM disks in (hyperdisk
        # storage pools).
        storage_pool = optional(string, "")

        # Customer-managed encryption key (CMEK) for the per-VM disks,
        # referenced as a GcpKmsKey. The Compute Engine service agent
        # (service-<project-number>@compute-system.iam.gserviceaccount.com)
        # must hold roles/cloudkms.cryptoKeyEncrypterDecrypter on the key.
        disk_encryption = optional(object({
          # The KMS key, referenced as a GcpKmsKey or a literal self link. The
          # service agent performing the operation needs
          # roles/cloudkms.cryptoKeyEncrypterDecrypter on it.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          kms_key = string

          # Service account used for the encryption/decryption request. When
          # omitted, the Compute Engine default service agent is used.
          kms_key_service_account = optional(string, "")
        }))

        # Decrypts the source image when it is itself CMEK-encrypted. Only
        # valid together with source_image.
        source_image_encryption = optional(object({
          # The KMS key, referenced as a GcpKmsKey or a literal self link. The
          # service agent performing the operation needs
          # roles/cloudkms.cryptoKeyEncrypterDecrypter on it.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          kms_key = string

          # Service account used for the encryption/decryption request. When
          # omitted, the Compute Engine default service agent is used.
          kms_key_service_account = optional(string, "")
        }))

        # Decrypts the source snapshot when it is itself CMEK-encrypted.
        # Only valid together with source_snapshot.
        source_snapshot_encryption = optional(object({
          # The KMS key, referenced as a GcpKmsKey or a literal self link. The
          # service agent performing the operation needs
          # roles/cloudkms.cryptoKeyEncrypterDecrypter on it.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          kms_key = string

          # Service account used for the encryption/decryption request. When
          # omitted, the Compute Engine default service agent is used.
          kms_key_service_account = optional(string, "")
        }))
      }))

      # Network interfaces for every VM. At least one is required; multiple
      # NICs must each attach to a different VPC network. Changing
      # interfaces rotates the template.
      network_interfaces = list(object({
        # VPC network for this interface, referenced as a GcpVpcNetwork.
        # Sufficient alone only for auto-mode VPCs; custom-mode VPCs need
        # subnetwork.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        network = optional(string, "")

        # Subnetwork for this interface, referenced as a GcpSubnetwork. The
        # subnetwork's region must contain the group's location.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        subnetwork = optional(string, "")

        # Project owning the subnetwork — set when attaching to a Shared VPC
        # host project's subnetwork from a service project.
        subnetwork_project = optional(string, "")

        # Static internal IP shared configuration is not meaningful for a
        # fleet (every VM needs its own address) — this field pins the
        # PRIMARY internal IP only for single-instance groups or specialized
        # setups. When omitted (the norm), GCP assigns each VM an ephemeral
        # internal IP from the subnetwork range.
        network_ip = optional(string, "")

        # External IPv4 access configs. Empty means no external IP (private
        # fleet — pair with Cloud NAT for egress; the secure default). GCP
        # supports at most one access config per interface.
        access_configs = optional(list(object({
          # Static external IP to pin — meaningful only for single-instance
          # groups (a fleet cannot share one IP; use stateful_external_ips for
          # per-instance identity). When omitted (the norm), each VM gets an
          # ephemeral external IP.
          nat_ip = optional(string, "")

          # Network service tier for the external IPs: "PREMIUM" (default;
          # Google's global backbone) or "STANDARD" (regional, cheaper).
          network_tier = optional(string, "")
        })), [])

        # External IPv6 access configs. Requires stack_type "IPV4_IPV6" and a
        # subnetwork with an external IPv6 range. At most one per interface.
        ipv6_access_configs = optional(list(object({
          # Network service tier for IPv6 traffic. Only "PREMIUM" is valid.
          network_tier = string
        })), [])

        # IP stack of the interface:
        #   ""            -- same as "IPV4_ONLY" (GCP default)
        #   "IPV4_ONLY"   -- IPv4 addresses only
        #   "IPV4_IPV6"   -- dual stack (subnetwork must have an IPv6 range)
        #   "IPV6_ONLY"   -- IPv6 only (supported on IPv6-enabled
        #                    subnetworks)
        stack_type = optional(string, "")

        # vNIC type: "" (GCP picks), "GVNIC" (recommended on modern machine
        # families; required for TIER_1 bandwidth), "VIRTIO_NET" (legacy), or
        # the RDMA types "IDPF", "MRDMA", "IRDMA" on specialized shapes.
        nic_type = optional(string, "")

        # Networking queue count for Rx and Tx (1-32). When omitted GCP sizes
        # queues from vCPU count.
        queue_count = optional(number)

        # Alias IP ranges served by this interface on every VM — per-VM
        # secondary ranges for multi-IP workloads.
        alias_ip_ranges = optional(list(object({
          # The alias range: a CIDR ("10.1.2.0/24"), a single IP ("10.1.2.3"),
          # or a netmask ("/24") to auto-allocate per VM from the range.
          ip_cidr_range = string

          # Secondary range name on the subnetwork to allocate from. When
          # omitted the primary range is used.
          subnetwork_range_name = optional(string, "")
        })), [])

        # URL of a Private Service Connect NETWORK ATTACHMENT this interface
        # connects to, in the form
        # "projects/{projectNumber}/regions/{region}/networkAttachments/{name}"
        # — connects the fleet into a producer's VPC. An attachment-only
        # interface is legal (no network or subnetwork).
        network_attachment = optional(string, "")

        # VLAN tag (2-255) making this a DYNAMIC network interface — a
        # sub-interface multiplexed onto a parent NIC.
        vlan = optional(number)

        # IGMP multicast query support on this interface:
        #   ""                    -- GCP default (disabled)
        #   "IGMP_QUERY_V2"       -- IGMPv2 queries enabled (multicast)
        #   "IGMP_QUERY_DISABLED" -- explicitly disabled
        igmp_query = optional(string, "")

        # Static INTERNAL IPv6 address for the interface — meaningful only
        # for single-instance groups (a fleet cannot share one address).
        # When omitted, GCP assigns per-VM addresses from the subnetwork's
        # internal IPv6 range.
        ipv6_address = optional(string, "")

        # Prefix length of the primary internal IPv6 range assigned to this
        # interface. When omitted, GCP assigns its default.
        internal_ipv6_prefix_length = optional(number)
      }))

      # Service account the VMs' workloads authenticate as. When omitted,
      # the Compute Engine default service account is used with its default
      # scopes — prefer a dedicated least-privilege account for production.
      # Changing it rotates the template.
      service_account = optional(object({
        # Service account email, referenced as a GcpServiceAccount or a
        # literal email.
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
      # Changing it rotates the template.
      scheduling = optional(object({
        # Provisioning model:
        #   ""                  -- same as "STANDARD"
        #   "STANDARD"          -- on-demand capacity
        #   "SPOT"              -- deeply discounted preemptible capacity;
        #                          GCP may reclaim VMs at any time (the group
        #                          recreates them per its repair policy)
        #   "FLEX_START"        -- discounted capacity with a flexible start
        #                          time (Dynamic Workload Scheduler); pairs
        #                          with resize_requests
        #   "RESERVATION_BOUND" -- runs only on capacity from one specific
        #                          reservation (pair with reservation_affinity
        #                          type SPECIFIC_RESERVATION)
        provisioning_model = optional(string, "")

        # Restart a VM automatically when Compute Engine (not a user) stops
        # it. Defaults to true for standard VMs; must be false (or unset) for
        # Spot/FLEX_START.
        automatic_restart = optional(bool)

        # Host maintenance behavior: "" (GCP default "MIGRATE"), "MIGRATE"
        # (live-migrate; zero downtime), or "TERMINATE" (stop during
        # maintenance — required for GPUs and confidential VMs).
        on_host_maintenance = optional(string, "")

        # What GCP does when a VM is reclaimed (Spot preemption, FLEX_START
        # expiry) or a run-duration limit fires: "STOP" (keep the stopped VM
        # and disks) or "DELETE" (remove the VM — the usual choice in a
        # managed group, which recreates capacity itself).
        instance_termination_action = optional(string, "")

        # Maximum run duration in seconds, after which
        # instance_termination_action is executed. The duration clock starts
        # at every VM start. Mutually exclusive with termination_time.
        max_run_duration_seconds = optional(number)

        # Absolute timestamp (RFC 3339) at which VMs are terminated.
        # Mutually exclusive with max_run_duration_seconds.
        termination_time = optional(string, "")

        # Discard local-SSD contents when a VM is stopped by a lifetime limit
        # (max_run_duration/termination_time) instead of preserving them.
        discard_local_ssds_on_stop = optional(bool)

        # Availability domain for spread-placement within the zone (used with
        # spread placement policies).
        availability_domain = optional(number)

        # Minimum vCPUs on the sole-tenant node the VMs can be scheduled
        # onto. Sole-tenancy only.
        min_node_cpus = optional(number)

        # Sole-tenant node affinities selecting which node groups the VMs may
        # run on. Setting any affinity places the fleet on sole-tenant
        # hardware.
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
        # when a host fails, in seconds, before falling back to default
        # recovery.
        local_ssd_recovery_timeout_seconds = optional(number)

        # How long Compute Engine waits, in seconds, before declaring a host
        # failed and starting host-error recovery for the VM on it. A lower
        # value recovers a hung host faster at the cost of more false
        # positives; leave unset for Compute Engine's default recovery timing.
        # Must be 90..330 in steps of 30 (90, 120, ..., 330).
        host_error_timeout_seconds = optional(number)
      }))

      # Shielded VM configuration (secure boot, vTPM, integrity
      # monitoring). Requires an image with Shielded VM support (all recent
      # Google-provided images qualify). Changing it rotates the template.
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

      # Confidential VM configuration — hardware memory encryption (AMD
      # SEV / SEV-SNP or Intel TDX). Requires a supported machine family
      # (e.g. N2D, C2D, C3) and scheduling.on_host_maintenance =
      # "TERMINATE". Changing it rotates the template.
      confidential_instance_config = optional(object({
        # Confidential computing technology:
        #   "SEV"     -- AMD Secure Encrypted Virtualization (N2D/C2D/C3D)
        #   "SEV_SNP" -- AMD SEV Secure Nested Paging (requires an AMD Milan+
        #                min_cpu_platform)
        #   "TDX"     -- Intel Trust Domain Extensions (C3)
        confidential_instance_type = string
      }))

      # Advanced machine features: nested virtualization, SMT control,
      # visible core count, UEFI networking, performance monitoring unit,
      # and turbo mode. Changing it rotates the template.
      advanced_machine_features = optional(object({
        # Expose nested virtualization support (VMX) to the guest — run VMs
        # inside the VMs.
        enable_nested_virtualization = optional(bool)

        # Threads per physical core: 1 disables simultaneous multithreading
        # (SMT) — common for licensing and security isolation; 2 is the
        # hardware default.
        threads_per_core = optional(number)

        # Number of physical cores exposed to the guest (core visibility for
        # per-core licensing). When unset all cores are visible.
        visible_core_count = optional(number)

        # Enable UEFI networking in the guest firmware.
        enable_uefi_networking = optional(bool)

        # Performance monitoring unit exposure level: "STANDARD", "ENHANCED",
        # or "ARCHITECTURAL".
        performance_monitoring_unit = optional(string, "")

        # Turbo frequency mode. "ALL_CORE_MAX" runs all cores at maximum
        # turbo frequency (supported machine families only).
        turbo_mode = optional(string, "")
      }))

      # GPU accelerator cards attached to every VM. Requires a GPU-capable
      # location and scheduling.on_host_maintenance = "TERMINATE".
      # Changing it rotates the template.
      guest_accelerators = optional(list(object({
        # Accelerator type available in the group's location, e.g.
        # "nvidia-tesla-t4", "nvidia-l4", "nvidia-a100-80gb".
        type = string

        # Number of cards of this type per VM.
        count = number
      })), [])

      # Reservation affinity — whether VMs consume capacity from any
      # matching reservation, a specific reservation, or none. Changing it
      # rotates the template.
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
      # "DEFAULT" is the standard cap. Changing it rotates the template.
      total_egress_bandwidth_tier = optional(string, "")

      # Custom metadata key/value pairs made available to the guest OS via
      # the metadata server. Well-known keys configure agents and features
      # (e.g. "enable-oslogin"). Changing metadata rotates the template.
      metadata = optional(map(string), {})

      # Startup script executed by the guest agent on every boot of every
      # VM. Maps to the metadata_startup_script surface, which keeps it
      # distinct from user metadata. Changing it rotates the template.
      startup_script = optional(string, "")

      # Network tags used by firewall rules and network routes to select
      # the group's VMs. Changing tags rotates the template.
      tags = optional(list(string), [])

      # User labels stamped onto every VM (merged with Planton attribution
      # labels, which win on key conflicts). The ONLY template surface GCP
      # allows to change in place — label edits do not rotate the template.
      labels = optional(map(string), {})

      # Resource Manager tags bound to the TEMPLATE for org-policy and IAM
      # conditions. Keys in the form "tagKeys/{id}", values
      # "tagValues/{id}". Changing them rotates the template.
      resource_manager_tags = optional(map(string), {})

      # Minimum CPU platform for the VMs, e.g. "Intel Ice Lake" or
      # "AMD Milan". Constrains scheduling to hosts with at least this
      # platform. Changing it rotates the template.
      min_cpu_platform = optional(string, "")

      # Allow sending/receiving packets with source or destination IPs
      # that do not match the VM's own addresses — required for VMs acting
      # as routers, NAT gateways, or VPN endpoints. Changing it rotates the
      # template.
      can_ip_forward = optional(bool, false)

      # Action GCP takes on the VMs when a Cloud KMS key protecting them is
      # revoked: "NONE" (default) or "STOP". Changing it rotates the
      # template.
      key_revocation_action_type = optional(string, "")

      # Self links of compute resource policies attached to every VM (e.g.
      # an instance schedule). GCP currently allows at most one policy per
      # instance. Changing it rotates the template.
      resource_policies = optional(list(string), [])

      # Managed workload identity for every VM in the group: a SPIFFE
      # identity issued to each instance (and, optionally, X.509 identity
      # certificates) so workloads authenticate to each other by identity
      # instead of shared secrets or network position. Part of the template,
      # so changing it rotates the template and rolls the group.
      workload_identity_config = optional(object({
        # The SPIFFE ID Compute Engine issues to each instance, e.g.
        # "spiffe://PROJECT.svc.id.goog/ns/NAMESPACE/sa/SERVICE_ACCOUNT" or a
        # workload-identity-pool identity of the form
        # "spiffe://POOL.global.PROJECT_NUMBER.workload.id.goog/ns/NS/sa/SA".
        identity = string

        # Whether Compute Engine also issues and rotates X.509 certificates
        # bound to the identity, made available on each VM for mutual TLS.
        identity_certificate_enabled = optional(bool, false)
      }))
    })

    # Named application versions for CANARY rollouts. When empty (the
    # normal case), the group runs one version on this kind's own
    # template. Add entries to split the fleet across templates — e.g.
    # the kind's template as the stable version plus an external
    # template URL pinned as a canary with a small target_size.
    versions = optional(list(object({
      # Name of the version (e.g. "stable", "canary") — appears in the
      # instance metadata of VMs created from it.
      version_name = optional(string, "")

      # Instance template for this version. Leave EMPTY to run this
      # kind's own template (the default and the normal case). Set a full
      # template self link URL to pin an EXTERNAL template — the canary
      # escape hatch for splitting the fleet across templates this kind
      # does not manage.
      template_self_link = optional(string, "")

      # Fixed number of instances running this version. Unset on the
      # stable version (it absorbs the remainder of the fleet).
      target_size_fixed = optional(number)

      # Percent of the fleet (0-100) running this version, rounded up.
      # Unset on the stable version (it absorbs the remainder).
      target_size_percent = optional(number)
    })), [])

    # Fixed number of running VMs in the group. Set this for manually
    # sized groups; leave it unset when the autoscaler manages the size
    # (the two are mutually exclusive — the autoscaler would fight a
    # fixed size on every apply). When neither target_size nor autoscaler
    # is set, the group is created with 0 instances.
    target_size = optional(number)

    # Named ports published by the group — the mechanism backend services
    # use to map a logical service name ("http") to a port number (8080)
    # on every VM. The same name must be used by the backend service's
    # port_name.
    named_ports = optional(list(object({
      # The port name backend services reference via port_name (e.g.
      # "http").
      name = string

      # The port number the service listens on (e.g. 8080).
      port = number
    })), [])

    # How the group rolls out template and configuration changes to
    # running VMs: automatically (PROACTIVE) or on-demand
    # (OPPORTUNISTIC), and within what surge/unavailability budget.
    # When omitted, the group manager applies changes opportunistically
    # with GCP's default budget.
    update_policy = optional(object({
      # The most disruptive action the rollout may take on its own:
      #   "NONE"    -- no automatic action (changes wait for manual
      #                refresh)
      #   "REFRESH" -- apply updates that need no restart
      #   "RESTART" -- stop/start instances to apply updates
      #   "REPLACE" -- recreate instances from the new template (the usual
      #                choice for template rotations)
      minimal_action = string

      # Rollout mode:
      #   "PROACTIVE"     -- the group rolls changes out automatically
      #                      within the surge/unavailability budget
      #   "OPPORTUNISTIC" -- changes apply only when instances are
      #                      recreated anyway (manual refresh, autoscaler
      #                      churn, repairs)
      type = string

      # Cap on how disruptive an individual update is allowed to be —
      # updates needing more disruption than this wait instead. Same value
      # set as minimal_action; defaults to "REPLACE" (no cap).
      most_disruptive_allowed_action = optional(string, "")

      # How replaced instances are recreated:
      #   ""           -- same as "SUBSTITUTE"
      #   "SUBSTITUTE" -- new instances get fresh random names (allows
      #                   surge; zero-unavailability rollouts possible)
      #   "RECREATE"   -- instance NAMES are preserved (stateful-friendly);
      #                   requires an unavailability budget above 0
      replacement_method = optional(string, "")

      # Extra instances the rollout may create above target_size (fixed
      # count). Higher surge = faster rollout, more temporary cost.
      # Regional groups accept ONLY 0 or a value >= the group's zone count
      # (live-verified 400: "Fixed updatePolicy.maxSurge for regional
      # managed instance group has to be either 0 or at least equal to the
      # number of zones" — a regional group spreads over 3 zones by
      # default, so 1 and 2 are rejected; use percent for finer budgets).
      max_surge_fixed = optional(number)

      # Extra instances the rollout may create above target_size, as a
      # percent of the group (0-100).
      max_surge_percent = optional(number)

      # Instances the rollout may take below target_size (fixed count).
      max_unavailable_fixed = optional(number)

      # Instances the rollout may take below target_size, as a percent of
      # the group (0-100).
      max_unavailable_percent = optional(number)

      # REGIONAL groups only: whether the group proactively redistributes
      # instances to keep the zone balance ("PROACTIVE", the GCP default)
      # or leaves them where they are ("NONE" — required for stateful
      # regional groups).
      instance_redistribution_type = optional(string, "")
    }))

    # Auto-healing: recreate VMs that fail an application-level health
    # check (not just VM liveness). Requires a GcpHealthCheck; the
    # initial delay gives freshly booted VMs time to become healthy
    # before repairs kick in.
    auto_healing = optional(object({
      # The health check that decides instance health, referenced as a
      # GcpHealthCheck or a literal self link. Use a health check tuned for
      # auto-healing (conservative thresholds) — aggressive LB health
      # checks cause repair storms.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      health_check = string

      # Seconds a freshly created VM gets to boot and become healthy before
      # auto-healing counts failures against it (0-3600). Size it to your
      # application's cold-start time — too short and the group repair-loops
      # healthy-but-slow instances.
      initial_delay_sec = number
    }))

    # Standby pool configuration: keep pre-created VMs SUSPENDED or
    # STOPPED so scale-outs resume them (seconds) instead of booting
    # from scratch (minutes).
    standby_policy = optional(object({
      # Seconds a newly created standby VM runs before being suspended or
      # stopped (0-3600) — time for boot and warmup so resume is instant.
      initial_delay_sec = optional(number)

      # How standby VMs are consumed:
      #   ""               -- same as "MANUAL" (GCP default)
      #   "MANUAL"         -- standby VMs activate only by explicit API
      #                       action
      #   "SCALE_OUT_POOL" -- scale-outs resume standby VMs first (the
      #                       fast-scale-out mode)
      mode = optional(string, "")
    }))

    # Target number of SUSPENDED VMs held in the standby pool
    # (suspended VMs keep memory state; fastest resume, disks and memory
    # keep billing).
    target_suspended_size = optional(number)

    # Target number of STOPPED VMs held in the standby pool (stopped VMs
    # keep only disks; cheaper than suspended, slower to start).
    target_stopped_size = optional(number)

    # Stateful persistent disks preserved across instance recreation and
    # updates — for VMs whose disks carry irreplaceable data (databases,
    # brokers). The device names must match disks defined in the
    # template.
    stateful_disks = optional(list(object({
      # Device name of the template disk to make stateful — must match a
      # device_name in the template's disks.
      device_name = string

      # What happens to the preserved disk when its instance is PERMANENTLY
      # deleted (group deletion, explicit instance delete — not repairs):
      #   ""                               -- same as "NEVER" (GCP default)
      #   "NEVER"                          -- the disk is kept
      #   "ON_PERMANENT_INSTANCE_DELETION" -- the disk is deleted with the
      #                                       instance
      delete_rule = optional(string, "")
    })), [])

    # Stateful EXTERNAL IPs preserved across instance recreation — each
    # VM keeps its public IP identity through repairs and updates.
    stateful_external_ips = optional(list(object({
      # Network interface name whose IP to preserve. When omitted, "nic0"
      # (the first interface) is used.
      interface_name = optional(string, "")

      # What happens to the preserved IP when its instance is PERMANENTLY
      # deleted:
      #   ""                               -- same as "NEVER" (GCP default)
      #   "NEVER"                          -- the address is kept
      #   "ON_PERMANENT_INSTANCE_DELETION" -- the address is released with
      #                                       the instance
      delete_rule = optional(string, "")
    })), [])

    # Stateful INTERNAL IPs preserved across instance recreation — each
    # VM keeps its private IP identity through repairs and updates.
    stateful_internal_ips = optional(list(object({
      # Network interface name whose IP to preserve. When omitted, "nic0"
      # (the first interface) is used.
      interface_name = optional(string, "")

      # What happens to the preserved IP when its instance is PERMANENTLY
      # deleted:
      #   ""                               -- same as "NEVER" (GCP default)
      #   "NEVER"                          -- the address is kept
      #   "ON_PERMANENT_INSTANCE_DELETION" -- the address is released with
      #                                       the instance
      delete_rule = optional(string, "")
    })), [])

    # What the group does when instances fail or need repair — the
    # repair-vs-do-nothing switches and the health-check failure action.
    instance_lifecycle_policy = optional(object({
      # What the group does with failed instances (crashed, preempted,
      # failed to start):
      #   ""           -- same as "REPAIR" (GCP default)
      #   "REPAIR"     -- recreate/restart failed instances automatically
      #   "DO_NOTHING" -- leave failed instances alone (manual operations
      #                   mode)
      default_action_on_failure = optional(string, "")

      # Whether a repair may apply the LATEST template version instead of
      # the instance's current one: "YES" or "NO" (GCP default NO —
      # repairs preserve the running version; YES turns repairs into
      # opportunistic update vectors).
      force_update_on_repair = optional(string, "")

      # What the group does when the auto-healing health check fails (only
      # meaningful with auto_healing configured):
      #   ""               -- same as "DEFAULT_ACTION" (GCP default)
      #   "DEFAULT_ACTION" -- follow default_action_on_failure
      #   "REPAIR"         -- repair on health-check failure even when
      #                       default_action_on_failure is DO_NOTHING
      #   "DO_NOTHING"     -- never repair on health-check failure
      on_failed_health_check = optional(string, "")

      # Whether a repair may recreate the instance in a DIFFERENT zone of a
      # regional group: "YES" or "NO" (GCP default NO — repairs stay
      # in-zone).
      on_repair_allow_changing_zone = optional(string, "")
    }))

    # Labels and metadata stamped onto ALL instances IN ADDITION to what
    # the template defines — changing these does NOT rotate the template;
    # the group patches running instances per update_policy instead. Use
    # for fleet-wide toggles that should not force template rotation.
    all_instances_config = optional(object({
      # Labels added to every instance (on top of template labels).
      labels = optional(map(string), {})

      # Metadata added to every instance (on top of template metadata).
      metadata = optional(map(string), {})
    }))

    # Pagination behavior of the group's listManagedInstances API:
    #   ""          -- same as "PAGELESS" (GCP default)
    #   "PAGELESS"  -- ignores pagination parameters (legacy behavior)
    #   "PAGINATED" -- respects maxResults/pageToken (use for very large
    #                  groups)
    list_managed_instances_results = optional(string, "")

    # Self link of a WORKLOAD resource policy applied to the group (e.g.
    # a high-throughput or placement workload policy). GCP accepts at
    # most one workload policy per group.
    workload_policy = optional(string, "")

    # Self links of legacy target pools (network load balancer) whose
    # member list the group manages. Modern L7/L4 load balancing uses
    # backend services pointed at the group's instance_group output
    # instead — target pools remain for the legacy NLB path.
    target_pools = optional(list(string), [])

    # Wait for all managed instances to reach the wait_for_instances_status
    # before the apply completes (and before dependent resources deploy).
    # Turns "the group exists" into "the fleet is actually up" — useful
    # when a chart deploys consumers right behind the group. Failed VMs
    # make the apply fail at timeout.
    wait_for_instances = optional(bool)

    # Which state wait_for_instances waits for:
    #   ""         -- same as "STABLE" (GCP default)
    #   "STABLE"   -- instances are running or repaired to running
    #   "UPDATED"  -- additionally, all instances are on the target
    #                 template version (rollouts fully converged)
    wait_for_instances_status = optional(string, "")

    # REGIONAL groups only: which zones VMs spread across and the target
    # distribution shape. When omitted, GCP spreads evenly across the
    # region's zones.
    distribution_policy = optional(object({
      # The zones instances may run in (e.g. ["us-central1-a",
      # "us-central1-f"]). When omitted, GCP uses the region's zones.
      # Immutable: changing zones replaces the group.
      zones = optional(list(string), [])

      # The distribution shape the group converges to (per the API's
      # documented set):
      #   ""                -- same as "EVEN" (GCP default)
      #   "EVEN"            -- equal instance counts across zones (highest
      #                        availability)
      #   "BALANCED"        -- spread across zones subject to capacity
      #   "ANY"             -- any zone with capacity (best for scarce
      #                        shapes)
      #   "ANY_SINGLE_ZONE" -- all instances in one zone GCP picks
      # The provider does not pre-validate this list — it is the API's
      # documented vocabulary, enforced here so typos fail at validate
      # time.
      target_shape = optional(string, "")
    }))

    # REGIONAL groups only: ranked machine-type alternatives the group
    # may fall back to when the primary shape is out of capacity —
    # useful for large fleets on capacity-constrained shapes.
    instance_flexibility_policy = optional(object({
      # Named, ranked machine-type selections. Lower rank is preferred.
      instance_selections = list(object({
        # Name of this selection (e.g. "primary", "fallback").
        name = string

        # Machine types in this selection (e.g. ["e2-standard-4",
        # "n2-standard-4"]).
        machine_types = list(string)

        # Preference rank — selections with lower rank are consumed first.
        rank = optional(number)
      }))
    }))

    # How the group creates VMs to reach its target size:
    #   ""           -- GCP default (individual creation)
    #   "INDIVIDUAL" -- create VMs one by one; partial success possible
    #   "BULK"       -- all-or-nothing atomic creation of the whole
    #                   target size
    # Immutable: changing it replaces the group.
    target_size_policy_mode = optional(string, "")

    # Autoscaler for the group — grows and shrinks the fleet between
    # min/max replicas from CPU, load-balancer serving capacity, custom
    # Cloud Monitoring metrics, or calendar schedules. Mutually
    # exclusive with target_size.
    autoscaler = optional(object({
      # Name of the autoscaler resource. When omitted, the group name is
      # used.
      autoscaler_name = optional(string, "")

      # Human-readable description of the autoscaler.
      description = optional(string, "")

      # Minimum number of replicas the autoscaler maintains (>= 0).
      min_replicas = optional(number, 0)

      # Maximum number of replicas the autoscaler may create.
      max_replicas = number

      # Seconds the autoscaler waits before collecting usage from a NEW
      # instance — covers boot and warmup so unreliable early samples do
      # not drive scaling. GCP default: 60.
      cooldown_period = optional(number)

      # Operating mode (the API's documented vocabulary; the provider
      # defaults it to "ON" and does not pre-validate the value):
      #   ""               -- same as "ON"
      #   "ON"             -- scale out and in
      #   "OFF"            -- autoscaler holds everything (group keeps its
      #                       current size)
      #   "ONLY_SCALE_OUT" -- grow but never shrink
      mode = optional(string, "")

      # Target CPU utilization as a fraction in (0, 1] — e.g. 0.6 keeps
      # average fleet CPU at 60%. The default signal when no other target
      # is set (GCP applies 0.6 when the policy has no targets at all).
      cpu_target = optional(number)

      # Predictive autoscaling for the CPU signal (the provider's own
      # documented values):
      #   ""                      -- same as "NONE"
      #   "NONE"                  -- react to real-time metrics only
      #   "OPTIMIZE_AVAILABILITY" -- learn daily/weekly patterns and scale
      #                              out AHEAD of anticipated demand
      cpu_predictive_method = optional(string, "")

      # Target backend utilization fraction (0, 1] of the load balancer's
      # configured serving capacity — scales the fleet to hold the
      # balancer's per-backend utilization at this level. Only meaningful
      # when the group serves an EXTERNAL HTTP(S) load balancer with
      # UTILIZATION balancing mode.
      load_balancing_target = optional(number)

      # Custom Cloud Monitoring metric signals.
      metrics = optional(list(object({
        # The metric identifier, e.g.
        # "pubsub.googleapis.com/subscription/num_undelivered_messages" or
        # "custom.googleapis.com/myapp/queue_depth". The metric must export
        # values for the fleet's instances (or be a per-group metric used
        # with filter + single_instance_assignment).
        name = string

        # Target value of the metric the autoscaler maintains per instance.
        # Pair with type to say how the metric is interpreted.
        target = optional(number)

        # How the metric's values are interpreted against target (the
        # provider validates this list):
        #   "GAUGE"             -- instantaneous value
        #   "DELTA_PER_SECOND"  -- rate per second
        #   "DELTA_PER_MINUTE"  -- rate per minute
        type = optional(string, "")

        # Monitoring filter expression selecting the metric's time series
        # (e.g. a per-group Pub/Sub subscription metric). GCP default:
        # "resource.type = gce_instance".
        filter = optional(string, "")

        # For per-GROUP workload metrics (queue depth, pending jobs): the
        # amount of work one instance handles. The autoscaler keeps
        # instances proportional to metric_value / single_instance_assignment.
        # Mutually exclusive with target/type.
        single_instance_assignment = optional(number)
      })), [])

      # Limits how fast the autoscaler SHRINKS the fleet after load drops —
      # the guard against cascading scale-in on temporary dips.
      scale_in_control = optional(object({
        # Max instances that may be removed within the trailing time window
        # (fixed count).
        max_scaled_in_replicas_fixed = optional(number)

        # Max instances that may be removed within the trailing time window,
        # as a percent of the group (0-100).
        max_scaled_in_replicas_percent = optional(number)

        # The trailing time window (seconds) the scale-in cap applies over.
        time_window_sec = optional(number)
      }))

      # Calendar-based capacity schedules (cron) that set a minimum replica
      # count for recurring time windows — e.g. business-hours capacity
      # floors. Metric-based scaling still adds capacity above the
      # schedule's floor.
      schedules = optional(list(object({
        # Name of the schedule (unique within the autoscaler).
        schedule_name = string

        # Cron expression for when the window STARTS (e.g. "0 8 * * MON-FRI"
        # for weekday mornings), interpreted in time_zone.
        schedule = string

        # How long the window lasts, in seconds. The provider documents a
        # minimum of 300.
        duration_sec = number

        # Minimum replicas the group holds during the window.
        min_required_replicas = number

        # Keep the schedule defined but inactive. GCP default: false.
        disabled = optional(bool)

        # IANA time zone the cron expression is evaluated in (e.g.
        # "America/New_York"). GCP default: "UTC".
        time_zone = optional(string, "")

        # Human-readable description of the schedule.
        description = optional(string, "")
      })), [])

      # Seconds of load stabilization the autoscaler considers before
      # scale-in decisions — effectively how long it "remembers" peak load.
      stabilization_period = optional(number)
    }))

    # Stateful per-instance overrides: pin a specific instance NAME to
    # preserved disks, IPs, and metadata. The group treats configured
    # instances as stateful — their identity and preserved state survive
    # recreation. The config name IS the instance name it applies to.
    per_instance_configs = optional(list(object({
      # The per-instance config's name — which IS the name of the managed
      # instance it applies to (for RECREATE-method groups the instance
      # names are "<base_instance_name>-<suffix>"; a config for a
      # not-yet-existing name creates an instance with that name).
      config_name = string

      # Preserved state pinned to the instance: disks, IPs, and metadata
      # that survive recreation.
      preserved_state = optional(object({
        # Preserved metadata key/value pairs pinned to the instance (merged
        # over template/all-instances metadata).
        metadata = optional(map(string), {})

        # Preserved disks pinned to the instance.
        disks = optional(list(object({
          # Device name the disk is exposed under on the instance — matches
          # the template's disk device_name when overriding a template disk.
          device_name = string

          # The disk to attach, referenced as a GcpComputeDisk or a literal
          # self link. The disk must live in the instance's zone.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          source = string

          # Attachment mode: "READ_WRITE" (GCP default) or "READ_ONLY".
          mode = optional(string, "")

          # What happens to the disk when the instance is PERMANENTLY deleted:
          #   ""                               -- same as "NEVER" (GCP default)
          #   "NEVER"                          -- the disk is kept
          #   "ON_PERMANENT_INSTANCE_DELETION" -- the disk is deleted with the
          #                                       instance
          delete_rule = optional(string, "")
        })), [])

        # Preserved EXTERNAL IPs pinned to the instance's interfaces.
        external_ips = optional(list(object({
          # Interface name the address is pinned to (e.g. "nic0").
          interface_name = string

          # The literal IP address to preserve, or a reference to a reserved
          # GcpAddress. When omitted, the instance's current address is
          # adopted as preserved state.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          address = optional(string, "")

          # What happens to the address when the instance is PERMANENTLY
          # deleted:
          #   ""                               -- same as "NEVER" (GCP default)
          #   "NEVER"                          -- the address is kept
          #   "ON_PERMANENT_INSTANCE_DELETION" -- the address is released with
          #                                       the instance
          auto_delete = optional(string, "")
        })), [])

        # Preserved INTERNAL IPs pinned to the instance's interfaces.
        internal_ips = optional(list(object({
          # Interface name the address is pinned to (e.g. "nic0").
          interface_name = string

          # The literal IP address to preserve, or a reference to a reserved
          # GcpAddress. When omitted, the instance's current address is
          # adopted as preserved state.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          address = optional(string, "")

          # What happens to the address when the instance is PERMANENTLY
          # deleted:
          #   ""                               -- same as "NEVER" (GCP default)
          #   "NEVER"                          -- the address is kept
          #   "ON_PERMANENT_INSTANCE_DELETION" -- the address is released with
          #                                       the instance
          auto_delete = optional(string, "")
        })), [])
      }))

      # The LEAST disruptive action the group may take to apply this
      # config to the instance (the update escalates to what the change
      # actually needs, but never below this): "NONE" (GCP default),
      # "REFRESH", "RESTART", or "REPLACE".
      minimal_action = optional(string, "")

      # The MOST disruptive action the group may take to apply this config
      # — updates needing more wait instead. "NONE", "REFRESH", "RESTART",
      # or "REPLACE" (GCP default).
      most_disruptive_allowed_action = optional(string, "")

      # When true, removing this config from the spec DELETES the managed
      # instance itself. Default false: the instance keeps running and
      # only its stateful config is removed (state application follows
      # remove_instance_state_on_destroy).
      remove_instance_on_destroy = optional(bool)

      # When true, removing this config also removes its preserved state
      # (disks per their delete rules, IPs, metadata) from the running
      # instance immediately. Default false: the instance keeps the state
      # until it is recreated. Irrelevant when remove_instance_on_destroy
      # deletes the instance altogether.
      remove_instance_state_on_destroy = optional(bool)
    })), [])

    # Queued one-shot capacity requests (Dynamic Workload Scheduler):
    # ask GCP to add N instances for a bounded duration when capacity
    # becomes available — the batch/HPC path for scarce shapes. Each
    # request is immutable once created; deleting an ACCEPTED request
    # cancels it.
    resize_requests = optional(list(object({
      # Name of the resize request (unique within the group).
      request_name = string

      # Human-readable description of the request.
      description = optional(string, "")

      # Number of instances to ADD to the group when capacity is granted.
      resize_by = number

      # How long the granted instances run before GCP reclaims them, in
      # seconds (10 minutes to 7 days: 600-604800 — the provider's own
      # documented bounds). Requires the template's
      # scheduling.provisioning_model to support bounded runs (FLEX_START).
      # When omitted, the request asks for unbounded capacity.
      requested_run_duration_seconds = optional(number)
    })), [])

    # Deletion policy — what happens to the group's resources when this
    # resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- every resource is deleted (VMs are terminated; disks
    #                follow their template auto_delete / stateful rules)
    #   "PREVENT" -- destroy FAILS before touching anything
    #   "ABANDON" -- resources are removed from management but keep
    #                running in GCP
    # Applies to every resource in the kind that supports it (group
    # manager, autoscaler, regional template, per-instance configs,
    # resize requests). The ZONAL instance template carries no deletion
    # policy in the provider — it is always deleted on destroy.
    deletion_policy = optional(string, "")
  })
}
