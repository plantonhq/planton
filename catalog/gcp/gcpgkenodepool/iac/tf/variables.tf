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
  description = "GcpGkeNodePool specification"
  type = object({
    # The GCP project the node pool is created in. Must be the parent
    # cluster's project — node pools cannot live in a different project than
    # their cluster. Accepts a literal project ID or a reference to a
    # GcpProject resource. If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the parent GKE cluster as created in GCP. Resolves from the
    # cluster's name output — the name GCP actually assigned — so it stays
    # correct even when the cluster's cloud name differs from its Planton
    # metadata.name. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    cluster_name = string

    # Location of the parent cluster — a region ("us-central1") for regional
    # clusters or a zone ("us-central1-a") for zonal ones. Must match the
    # cluster's own location; resolves from the cluster's location output by
    # default. Immutable. For a regional cluster, size limits in autoscaling
    # are PER ZONE (see autoscaling), and the pool gets one managed instance
    # group per zone.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    location = string

    # Name of the node pool in GKE. Immutable. If not specified, defaults to
    # metadata.name. Must be 1-40 characters: lowercase letters, digits, and
    # hyphens; starting with a letter and ending with a letter or digit.
    # Example: "general-pool", "spot-batch", "gpu-a100"
    node_pool_name = optional(string, "")

    # Prefix for a GENERATED pool name — GKE appends a random suffix, so
    # every replacement pool gets a fresh unique name (the
    # create-before-destroy pattern for pools that must be swapped without
    # a name collision). Mutually exclusive with node_pool_name; when both
    # are empty, node_pool_name defaults to metadata.name. Immutable.
    name_prefix = optional(string, "")

    # Zones this pool's nodes run in, e.g. ["us-central1-a", "us-central1-b"].
    # Must be within the cluster's region. If unspecified, the cluster-level
    # node_locations apply (all zones in the region for a regional cluster).
    # Mutable. Narrowing a pool to fewer zones cuts cost and inter-zone
    # traffic for workloads that tolerate zonal risk.
    node_locations = optional(list(string), [])

    # Kubernetes version for the nodes. If empty (recommended), GKE picks the
    # version and auto-upgrade keeps it current. Setting an explicit version
    # while management.auto_upgrade is on makes the two fight — pin a version
    # only with auto-upgrade off, and expect to own upgrades manually from
    # then on.
    version = optional(string, "")

    # Maximum pods per node in this pool (8-256), overriding the cluster's
    # default (110). Lower values shrink the per-node pod CIDR slice so the
    # pod range stretches across more nodes. Immutable. Only effective on
    # VPC-native clusters (which every Planton GKE cluster is).
    max_pods_per_node = optional(number)

    # Number of nodes the pool starts with when autoscaling manages the size
    # (per zone for regional clusters). Only meaningful with autoscaling —
    # a fixed-size pool's size IS node_count. Immutable: changing it forces
    # pool recreation, so set it once and let the autoscaler own the size
    # from then on.
    initial_node_count = optional(number)

    # Fixed number of nodes (per zone for regional clusters). The pool
    # stays at this size until you change it.
    node_count = optional(number)

    # Cluster-autoscaler management of the pool size between bounds.
    autoscaling = optional(object({
      # Minimum nodes PER ZONE. 0 allows scale-to-zero — the pattern for Spot
      # and GPU pools that should cost nothing while idle.
      min_nodes = optional(number)

      # Maximum nodes PER ZONE. A regional cluster in 3 zones with max_nodes=4
      # can reach 12 nodes.
      max_nodes = optional(number)

      # Minimum nodes across ALL zones (total addressing mode).
      total_min_nodes = optional(number)

      # Maximum nodes across ALL zones (total addressing mode) — an absolute
      # cost cap regardless of zone spread.
      total_max_nodes = optional(number)

      # Scale-up algorithm: BALANCED spreads new nodes to even out zone sizes;
      # ANY prioritizes unused reservations and reduces Spot preemption risk —
      # prefer ANY for Spot pools.
      location_policy = optional(string, "")
    }))

    # Auto-repair and auto-upgrade. If omitted, both default to true — GKE's
    # own defaults and the posture release channels expect.
    management = optional(object({
      # Automatically repair nodes that fail health checks. Keep on; a pool of
      # broken nodes heals itself instead of paging you.
      auto_repair = optional(bool)

      # Automatically upgrade node Kubernetes versions to track the control
      # plane. Keep on unless you pin `version` and own upgrades manually
      # (required for pools on a release channel).
      auto_upgrade = optional(bool)
    }))

    # How node upgrades roll through the pool: surge (default — add
    # max_surge nodes, drain max_unavailable at a time) or blue-green
    # (provision a full green pool, shift workloads, soak, delete blue).
    # If omitted, GKE's default surge settings (max_surge=1,
    # max_unavailable=0) apply.
    upgrade_settings = optional(object({
      # Additional nodes added during a surge upgrade (0 or more). Higher means
      # faster upgrades at temporary extra cost. max_surge + max_unavailable
      # must be at least 1 and at most 20 combined.
      max_surge = optional(number)

      # Nodes that may be simultaneously unavailable during a surge upgrade.
      # Higher trades workload disruption for speed.
      max_unavailable = optional(number)

      # SURGE (default): rolling replacement, cheapest. BLUE_GREEN: provision a
      # complete new node set, migrate, soak, then delete the old one — safest
      # rollback story, double capacity while in flight.
      strategy = optional(string, "")

      # Blue-green rollout pacing. Only with strategy BLUE_GREEN.
      blue_green_settings = optional(object({
        # How the blue pool drains, batch by batch.
        standard_rollout_policy = object({
          # Fraction of blue nodes drained per batch, 0.0-1.0.
          batch_percentage = optional(number)

          # Number of blue nodes drained per batch.
          batch_node_count = optional(number)

          # Soak time after each batch drains before the next starts. Duration in
          # seconds format, e.g. "600s".
          batch_soak_duration = optional(string, "")
        })

        # Time after the entire blue pool is drained before it is deleted —
        # the rollback window. Duration in seconds format, e.g. "3600s".
        node_pool_soak_duration = optional(string, "")
      }))
    }))

    # Compact placement: co-locates nodes physically for low inter-node
    # latency (tightly-coupled HPC/ML workloads) — and carries the TPU
    # topology for TPU pools. Immutable.
    placement_policy = optional(object({
      # COMPACT places nodes close together for minimal inter-node latency —
      # for tightly-coupled HPC and multi-node ML training. Requires machine
      # families that support compact placement (C2/C3/A2/A3...).
      type = string

      # Optional user-supplied compute resource policy to place nodes under.
      # Must be in the same project and region as the pool. When empty, GKE
      # creates and owns the placement policy.
      policy_name = optional(string, "")

      # TPU placement topology, e.g. "2x2x2" — TPU pools only.
      # https://cloud.google.com/kubernetes-engine/docs/concepts/plan-tpus#topology
      tpu_topology = optional(string, "")
    }))

    # Nodes are obtainable only through the ProvisioningRequest API (Dynamic
    # Workload Scheduler queued provisioning) — the pool provisions capacity
    # in whole batches when it becomes available, for large atomic workloads
    # like multi-node training jobs. Immutable.
    queued_provisioning_enabled = optional(bool, false)

    # Pool-level networking overrides (pod range, private nodes, network
    # performance). If omitted, the cluster-level defaults apply.
    network_config = optional(object({
      # Create a NEW secondary range for this pool's pods (named by pod_range,
      # sized by pod_ipv4_cidr_block) instead of drawing from the cluster's pod
      # range. Immutable. Dedicated pod ranges isolate address planning per
      # pool — useful when one pool's churn would exhaust the shared range.
      create_pod_range = optional(bool, false)

      # The secondary range for this pool's pod IPs. With create_pod_range,
      # the name given to the new range; without it, the name of an EXISTING
      # secondary range on the cluster's subnetwork. Immutable.
      pod_range = optional(string, "")

      # CIDR (e.g. "10.96.0.0/14") or netmask size (e.g. "/14") for the new pod
      # range when create_pod_range is set; empty lets GKE choose. Immutable.
      pod_ipv4_cidr_block = optional(string, "")

      # Whether this pool's nodes get only internal IPs, overriding the
      # cluster-level private-nodes setting. Unset inherits from the cluster.
      enable_private_nodes = optional(bool)

      # Network bandwidth tier: TIER_1 unlocks up to 100 Gbps total egress on
      # supported machine families (N2/N2D/C2/C3...). Requires gVNIC on this
      # pool (node_config.gvnic_enabled).
      total_egress_bandwidth_tier = optional(string, "")

      # Disables the pod CIDR overprovisioning for this pool (GKE normally
      # doubles the pod range slice per node). Only relevant with a dedicated
      # pod range. Immutable.
      pod_cidr_overprovision_disabled = optional(bool, false)

      # Places this pool's nodes on a DIFFERENT subnetwork than the cluster's
      # default node subnetwork (same VPC). Accepts a subnetwork self link or
      # a reference to a GcpSubnetwork resource. Immutable. Used to give pools
      # their own address planning or firewall scope.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnetwork = optional(string, "")

      # Accelerator network profile for the pool (e.g. RDMA-capable profiles
      # on GPU supercomputer shapes). GKE validates the profile against the
      # machine family at apply. Immutable.
      accelerator_network_profile = optional(string, "")

      # Additional node network interfaces (multi-networking): each entry
      # attaches every node to one more VPC network/subnetwork. Requires the
      # cluster's enable_multi_networking. Immutable.
      additional_node_networks = optional(list(object({
        # The VPC network to attach. Accepts a network name/self link or a
        # reference to a GcpVpcNetwork resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        network = string

        # The subnetwork on that network the interface draws its IP from.
        # Accepts a name/self link or a reference to a GcpSubnetwork resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        subnetwork = string
      })), [])

      # Additional pod ranges (multi-networking): each entry makes a
      # secondary range on a subnetwork available to this pool's pods beyond
      # the primary pod range. Requires the cluster's
      # enable_multi_networking. Immutable.
      additional_pod_networks = optional(list(object({
        # The subnetwork holding the secondary range. Accepts a name/self link
        # or a reference to a GcpSubnetwork resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        subnetwork = optional(string, "")

        # Name of the secondary range on that subnetwork used for pod IPs.
        secondary_pod_range = string

        # Maximum pods per node drawing from this range (8-256).
        max_pods_per_node = optional(number)
      })), [])
    }))

    # The node VM configuration: machine type, disks, identity, scheduling
    # constraints, accelerators, security posture, and kubelet/OS tuning.
    # If omitted entirely, GKE defaults apply (e2-medium, 100 GB pd-balanced,
    # Container-Optimized OS, Compute Engine default service account).
    node_config = optional(object({
      # Compute Engine machine type, e.g. "e2-medium", "n2-standard-8",
      # "a2-highgpu-1g". Defaults to e2-medium (2 vCPU, 4 GB) — fine for
      # sandboxes, undersized for real workloads. Immutable (changing it
      # replaces the pool's nodes).
      machine_type = optional(string)

      # Boot disk size in GB per node (min 10). GKE's default is 100.
      disk_size_gb = optional(number)

      # Boot disk type. pd-balanced (GKE's current default) suits most pools;
      # pd-ssd for I/O-heavy node-local work; hyperdisk-balanced on machine
      # families that require it (C3/C3D and newer).
      disk_type = optional(string, "")

      # Node OS image. COS_CONTAINERD (Container-Optimized OS, the default and
      # GKE's recommendation), UBUNTU_CONTAINERD (when you need Ubuntu
      # packages/kernel modules), or WINDOWS_LTSC_CONTAINERD for Windows
      # workloads.
      image_type = optional(string)

      # Service account the node VMs run as. Accepts an email literal or a
      # reference to a GcpServiceAccount resource. Defaults to the Compute
      # Engine default SA — create a minimal dedicated SA for production and
      # grant workload permissions through Workload Identity instead of node
      # scopes.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      service_account = optional(string, "")

      # OAuth scopes on the node VMs. Empty applies GKE's defaults
      # (devstorage.read_only, logging.write, monitoring). With Workload
      # Identity (the Planton cluster default), workload permissions come from
      # IAM on Kubernetes service accounts — node scopes only gate node-level
      # agents, so the defaults are usually right. One node-level agent that
      # matters: the kubelet pulls images from Artifact Registry with the node
      # service account under devstorage.read_only, so a private image in a
      # repository that account is granted on needs no pull secret at all.
      # Dropping that scope, or a repository the account is not granted on, is
      # why a pod would need a login declared on the workload instead.
      oauth_scopes = optional(list(string), [])

      # Kubernetes labels applied to every node in the pool — what nodeSelector
      # and affinity rules match on. Example: {"workload-class": "batch"}.
      labels = optional(map(string), {})

      # GCE resource labels on the node VMs (cloud billing/inventory labels,
      # not Kubernetes labels). Merged with the standard platform labels;
      # platform attribution keys win on conflict.
      resource_labels = optional(map(string), {})

      # GCE network tags on the node VMs — what VPC firewall rules match.
      # GKE adds its own cluster tag automatically; entries here are additive.
      tags = optional(list(string), [])

      # GCE instance metadata key/value pairs. GKE requires
      # "disable-legacy-endpoints" = "true" and both engines enforce it
      # beneath any entries set here. Immutable.
      metadata = optional(map(string), {})

      # Kubernetes taints applied to every node — the scheduling fence that
      # keeps general workloads off special-purpose pools. Pair each taint
      # with a matching toleration on the intended workloads. GPU pools get
      # an automatic nvidia.com/gpu taint from GKE.
      taints = optional(list(object({
        # Taint key, e.g. "workload-class".
        key = string

        # Taint value, e.g. "batch".
        value = string

        # NO_SCHEDULE fences new pods without a toleration; PREFER_NO_SCHEDULE
        # is advisory; NO_EXECUTE also evicts running pods without a toleration.
        effect = string
      })), [])

      # Spot VMs: deeply discounted (60-91%) capacity with no availability
      # guarantee — nodes can be preempted with 30s notice. The current model
      # (no 24-hour max lifetime; replaces preemptible). Pair with
      # scale-to-zero autoscaling, ANY location policy, and taints so only
      # fault-tolerant workloads land here. Immutable.
      spot = optional(bool, false)

      # Legacy preemptible VMs (24-hour max lifetime). Prefer spot for new
      # pools. Immutable.
      preemptible = optional(bool, false)

      # GPU accelerators attached to every node. The machine type must support
      # the accelerator (or be an accelerator-optimized family like A2/A3/G2,
      # where the GPU is implied by the machine type and this block must match
      # it). GKE taints GPU nodes automatically.
      guest_accelerators = optional(list(object({
        # Accelerator type resource name, e.g. "nvidia-tesla-t4",
        # "nvidia-l4", "nvidia-a100-80gb". Must be available in the pool's
        # zones.
        type = string

        # Number of accelerator cards per node.
        count = optional(number, 0)

        # NVIDIA MIG partition size, e.g. "1g.5gb" — slices one physical GPU
        # into isolated instances (A100/H100 families).
        gpu_partition_size = optional(string, "")

        # Driver installation: DEFAULT (GKE installs the default driver
        # version), LATEST (newest available; COS only), or
        # INSTALLATION_DISABLED (bring your own DaemonSet). If omitted on GKE
        # 1.30.1+, DEFAULT applies.
        gpu_driver_version = optional(string, "")

        # GPU sharing: lets multiple pods share one physical GPU.
        gpu_sharing_config = optional(object({
          # GPU_TIME_SHARING context-switches the GPU between pods; MPS (NVIDIA
          # Multi-Process Service) runs them concurrently with resource limits.
          gpu_sharing_strategy = string

          # Maximum pods sharing each physical GPU.
          max_shared_clients_per_gpu = optional(number, 0)
        }))
      })), [])

      # Shielded VM options. GKE's defaults: secure boot off (to tolerate
      # unsigned third-party kernel modules), integrity monitoring on. Enable
      # secure boot unless a workload loads unsigned modules. Immutable.
      shielded_instance_config = optional(object({
        # Verify boot components against a signature baseline. GCP default
        # false — because unsigned third-party kernel modules fail secure boot.
        # Turn it on unless you load such modules.
        enable_secure_boot = optional(bool, false)

        # Monitor and attest boot integrity at runtime (GCP default true).
        enable_integrity_monitoring = optional(bool)
      }))

      # Confidential GKE nodes: hardware memory encryption (AMD SEV / Intel
      # TDX) for the node VMs. Requires a supporting machine family (N2D/C2D/
      # C3D...). Immutable.
      confidential_nodes = optional(object({
        # Whether confidential nodes are enabled for this pool.
        enabled = optional(bool, false)

        # Confidential computing technology: SEV (AMD, the common choice),
        # SEV_SNP, or TDX (Intel). Empty lets GCP choose for the machine family.
        confidential_instance_type = optional(string, "")
      }))

      # Minimum CPU platform, e.g. "Intel Ice Lake" — pins scheduling to that
      # CPU generation or newer for instruction-set or performance floors.
      # Immutable.
      min_cpu_platform = optional(string, "")

      # Local SSDs attached to each node for scratch I/O, automatically
      # formatted and mounted by GKE (SCSI interface; legacy knob). For NVMe,
      # prefer ephemeral_storage_local_ssd or local_nvme_ssd_block. Immutable.
      local_ssd_count = optional(number)

      # Back ephemeral storage (emptyDir, container layers, logs) with local
      # NVMe SSDs instead of the boot disk — the biggest node-side I/O win for
      # churn-heavy workloads. Immutable.
      ephemeral_storage_local_ssd = optional(object({
        # Number of local SSDs backing ephemeral storage. Each is 375 GB (or
        # 3000 GB on Z3); count must match what the machine type supports.
        local_ssd_count = optional(number, 0)

        # Of those, SSDs dedicated to GKE Data Cache (read caching for
        # persistent volumes).
        data_cache_count = optional(number)
      }))

      # Attach raw-block local NVMe SSDs for workloads that manage their own
      # filesystem (databases with direct block access). Immutable.
      local_nvme_ssd_block = optional(object({
        # Number of raw-block local NVMe SSDs (375 GB each).
        local_ssd_count = optional(number, 0)
      }))

      # Image streaming (GCFS): containers start before the full image is
      # pulled, pulling data on demand — large-image pools (ML frameworks)
      # start minutes faster. Requires Container-Optimized OS.
      gcfs_enabled = optional(bool, false)

      # gVNIC (Google Virtual NIC): higher-throughput networking than virtio;
      # required for TIER_1 bandwidth and 100+ Gbps machine shapes. Immutable.
      gvnic_enabled = optional(bool, false)

      # NCCL Fast Socket: optimizes multi-node collective communication for
      # distributed GPU training. Requires gvnic_enabled.
      fast_socket_enabled = optional(bool, false)

      # Customer-managed encryption key (CMEK) for the node boot disks.
      # Accepts a full crypto key path or a reference to a GcpKmsKey resource.
      # The Compute Engine service agent needs Encrypter/Decrypter on the key.
      # Immutable.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      boot_disk_kms_key = optional(string, "")

      # How workloads see instance metadata: GKE_METADATA runs the GKE
      # metadata server per node (required for Workload Identity — the default
      # on WI clusters and the right answer); GCE_METADATA exposes the raw VM
      # metadata (legacy, leaks node credentials to pods).
      workload_metadata_mode = optional(string, "")

      # Consume Compute Engine reservations: committed capacity for the pool.
      # ANY_RESERVATION uses any matching reservation; SPECIFIC_RESERVATION
      # targets one by name (key "compute.googleapis.com/reservation-name",
      # values [name]); NO_RESERVATION opts out. Immutable.
      reservation_affinity = optional(object({
        # NO_RESERVATION opts out; ANY_RESERVATION consumes any matching
        # reservation; SPECIFIC_RESERVATION targets one by name;
        # ANY_RESERVATION_THEN_FAIL consumes matching reservations and fails
        # scale-up (instead of falling back to on-demand) when none is
        # available.
        consume_reservation_type = string

        # Reservation label key — "compute.googleapis.com/reservation-name" for
        # SPECIFIC_RESERVATION.
        key = optional(string, "")

        # Reservation label values — the reservation name(s).
        values = optional(list(string), [])
      }))

      # Secondary boot disks that preload container images or data onto every
      # node — cold-start acceleration for very large images. Immutable.
      secondary_boot_disks = optional(list(object({
        # Disk image to create the secondary boot disk from (a prepared image
        # containing the container images/data to preload).
        disk_image = string

        # CONTAINER_IMAGE_CACHE serves preloaded container images to the
        # container runtime. Empty attaches the disk without special handling.
        mode = optional(string, "")
      })), [])

      # Kubelet tuning: CPU management, PID limits, log rotation, image GC.
      # Only set what you need; unset fields keep GKE defaults.
      kubelet_config = optional(object({
        # CPU management: "static" gives Guaranteed-QoS pods exclusive cores
        # (latency-sensitive workloads); "none" (default) shares cores.
        cpu_manager_policy = optional(string, "")

        # Enforce CPU CFS quota for containers with CPU limits. Disabling trades
        # throttling for potential node CPU contention.
        cpu_cfs_quota = optional(bool)

        # CFS quota period, e.g. "100ms" (kubelet default) — shorter periods
        # smooth throttling for latency-sensitive workloads.
        cpu_cfs_quota_period = optional(string, "")

        # Maximum processes per pod — a fork-bomb fence for multi-tenant pools.
        pod_pids_limit = optional(number)

        # The kubelet's insecure read-only port 10255: FALSE closes it (the
        # hardened posture new clusters default to); TRUE keeps it open for
        # legacy monitoring agents that still scrape it.
        insecure_kubelet_readonly_port_enabled = optional(string, "")

        # Parallel image pulls (kubelet default serializes fewer) — speeds up
        # busy nodes that churn many images.
        max_parallel_image_pulls = optional(number)

        # Container log file size before rotation, e.g. "10Mi" (10Mi-500Mi).
        container_log_max_size = optional(string, "")

        # Rotated container log files kept per container (2-10).
        container_log_max_files = optional(number)

        # Disk usage percent below which image GC never runs (must be lower
        # than the high threshold).
        image_gc_low_threshold_percent = optional(number)

        # Disk usage percent above which image GC always runs.
        image_gc_high_threshold_percent = optional(number)

        # Minimum age an unused image must reach before GC may remove it,
        # seconds format, e.g. "120s".
        image_minimum_gc_age = optional(string, "")

        # Maximum age an unused image may reach before GC removes it regardless
        # of disk pressure, seconds format, e.g. "86400s".
        image_maximum_gc_age = optional(string, "")

        # Sysctl patterns pods may set as unsafe sysctls (e.g. "net.*",
        # "kernel.shm*") — gate carefully: unsafe sysctls affect the whole
        # node, not just the pod that sets them.
        allowed_unsafe_sysctls = optional(list(string), [])

        # Maximum seconds the kubelet grants a pod to terminate during
        # soft-eviction (caps the pod's own grace period in that path).
        eviction_max_pod_grace_period_seconds = optional(number)

        # Kill only the offending process instead of the whole container on
        # OOM — for containers running multiple processes where one leaking
        # process should not take down its siblings.
        single_process_oom_kill = optional(bool)

        # Soft eviction thresholds: the kubelet starts graceful pod eviction
        # when a signal stays past its threshold for the paired grace period.
        eviction_soft = optional(object({
          # Threshold for memory.available.
          memory_available = optional(string, "")

          # Threshold for nodefs.available (the filesystem backing pod volumes
          # and logs).
          nodefs_available = optional(string, "")

          # Threshold for nodefs.inodesFree.
          nodefs_inodes_free = optional(string, "")

          # Threshold for imagefs.available (the filesystem backing container
          # images and writable layers).
          imagefs_available = optional(string, "")

          # Threshold for imagefs.inodesFree.
          imagefs_inodes_free = optional(string, "")

          # Threshold for pid.available.
          pid_available = optional(string, "")
        }))

        # Grace periods paired with eviction_soft: how long a signal must stay
        # past its threshold before eviction starts (durations like "90s").
        eviction_soft_grace_period = optional(object({
          # Grace period for memory.available.
          memory_available = optional(string, "")

          # Grace period for nodefs.available.
          nodefs_available = optional(string, "")

          # Grace period for nodefs.inodesFree.
          nodefs_inodes_free = optional(string, "")

          # Grace period for imagefs.available.
          imagefs_available = optional(string, "")

          # Grace period for imagefs.inodesFree.
          imagefs_inodes_free = optional(string, "")

          # Grace period for pid.available.
          pid_available = optional(string, "")
        }))

        # Minimum amounts reclaimed per eviction: prevents thrashing by making
        # each eviction free at least this much of the signal's resource.
        eviction_minimum_reclaim = optional(object({
          # Minimum reclaim for memory.available, percentage like "10%".
          memory_available = optional(string, "")

          # Minimum reclaim for nodefs.available, percentage like "10%".
          nodefs_available = optional(string, "")

          # Minimum reclaim for nodefs.inodesFree, percentage like "10%".
          nodefs_inodes_free = optional(string, "")

          # Minimum reclaim for imagefs.available, percentage like "10%".
          imagefs_available = optional(string, "")

          # Minimum reclaim for imagefs.inodesFree, percentage like "10%".
          imagefs_inodes_free = optional(string, "")

          # Minimum reclaim for pid.available, percentage like "10%".
          pid_available = optional(string, "")
        }))

        # Caps the exponential backoff for restarting crashed containers —
        # lower caps recover crash-looping containers faster at the cost of
        # more restart churn.
        crash_loop_back_off = optional(object({
          # Maximum backoff delay between restarts of a crashing container,
          # seconds format (e.g. "300s"; kubelet default caps at 300s).
          max_container_restart_period = optional(string, "")
        }))

        # Kubernetes Memory Manager policy: Static reserves exclusive NUMA
        # memory for Guaranteed-QoS pods; None (default) does not.
        memory_manager = optional(object({
          # "Static" reserves exclusive NUMA-aligned memory for Guaranteed-QoS
          # pods; "None" disables the manager. Capitalized, per the kubelet's
          # own policy names.
          policy = optional(string, "")
        }))

        # Kubernetes Topology Manager: aligns CPU, memory, and device (GPU)
        # NUMA placement per pod or per container — for latency-critical and
        # HPC workloads that suffer on cross-NUMA access.
        topology_manager = optional(object({
          # Alignment policy: none (default), best-effort, restricted, or
          # single-numa-node (strictest — pods failing alignment are rejected).
          policy = optional(string, "")

          # Alignment scope: container (default; each container aligned
          # independently) or pod (all containers of a pod share one alignment).
          scope = optional(string, "")
        }))

        # Graceful node shutdown: the total time, in seconds, a node delays its
        # shutdown so every pod (critical and non-critical) can terminate
        # cleanly when the VM is reclaimed. Only configurable on Spot or
        # preemptible pools (the ones that get reclaimed). Between 10 and
        # 10000. Leave unset for GKE's default; sent only when set because the
        # API fills the value itself.
        shutdown_grace_period_seconds = optional(number)

        # The portion of shutdown_grace_period_seconds reserved for critical
        # pods (system-node-critical and system-cluster-critical priority
        # classes) after ordinary pods have been given their share. Must not
        # exceed shutdown_grace_period_seconds. Spot or preemptible pools only.
        # Sent only when set because the API fills the value itself.
        shutdown_grace_period_critical_pods_seconds = optional(number)
      }))

      # Linux node OS tuning: sysctls, cgroup mode, hugepages.
      linux_node_config = optional(object({
        # Sysctls applied to every node, e.g.
        # {"net.core.somaxconn": "4096"} — only GKE's allowlisted keys.
        sysctls = optional(map(string), {})

        # Container runtime cgroup mode: CGROUP_MODE_V2 (the modern default on
        # current GKE versions) or CGROUP_MODE_V1 for workloads that still need
        # v1.
        cgroup_mode = optional(string, "")

        # Hugepage pre-allocation for DPDK/database workloads.
        hugepages_config = optional(object({
          # Number of 2MB hugepages.
          hugepage_size_2m = optional(number)

          # Number of 1GB hugepages.
          hugepage_size_1g = optional(number)
        }))

        # Kernel transparent hugepage mode for anonymous memory:
        # ALWAYS, MADVISE (only regions that request it), or NEVER.
        transparent_hugepage_enabled = optional(string, "")

        # Kernel defrag behavior when allocating transparent hugepages:
        # ALWAYS (stall to reclaim), DEFER (kick background reclaim),
        # DEFER_WITH_MADVISE, MADVISE, or NEVER.
        transparent_hugepage_defrag = optional(string, "")

        # Signed-kernel-module enforcement: ENFORCE_SIGNED_MODULES rejects
        # unsigned module loads (the hardened posture);
        # DO_NOT_ENFORCE_SIGNED_MODULES allows them.
        node_kernel_module_loading_policy = optional(string, "")

        # PTP/KVM paravirtual clock sync for sub-millisecond time accuracy —
        # for workloads needing precise cross-node timestamps (trading,
        # distributed tracing at fine granularity).
        enable_ptp_kvm_time_sync = optional(bool)

        # Swap on the nodes (Kubernetes swap support): sizing profile plus
        # encryption. Swap trades OOM kills for latency under memory pressure;
        # pair with kubelet eviction tuning.
        swap_config = optional(object({
          # Whether swap is enabled on the nodes.
          enabled = optional(bool)

          # Swap carved from the boot disk.
          boot_disk_profile = optional(object({
            # Absolute swap size in GiB.
            swap_size_gib = optional(number)

            # Swap size as a percentage of the backing storage (1-100).
            swap_size_percent = optional(number)
          }))

          # Swap on local SSDs dedicated entirely to swap.
          dedicated_local_ssd_profile = optional(object({
            # Number of local SSDs dedicated to swap.
            disk_count = optional(number)
          }))

          # Swap carved from the ephemeral-storage local SSDs.
          ephemeral_local_ssd_profile = optional(object({
            # Absolute swap size in GiB.
            swap_size_gib = optional(number)

            # Swap size as a percentage of the backing storage (1-100).
            swap_size_percent = optional(number)
          }))

          # Swap encryption (on by default; disabling trades confidentiality for
          # a little throughput).
          encryption_config = optional(object({
            # Set true to DISABLE swap encryption (encrypted by default).
            disabled = optional(bool)
          }))
        }))

        # A custom initialization script GKE runs on every node at boot,
        # before the node joins the cluster — for host-level setup that no
        # DaemonSet can do (kernel parameters that need a reboot-free apply,
        # vendor agents, custom certificates). Sourced from Cloud Storage or
        # Secret Manager; exactly one source.
        custom_node_init = optional(object({
          # Cloud Storage object holding the script, e.g.
          # "gs://my-bucket/node-init.sh". The node's service account needs read
          # access to the object.
          gcs_uri = optional(string, "")

          # Pin the Cloud Storage object to one generation so a later upload
          # does not silently change what new nodes run. Leave unset to always
          # fetch the current object; sent only when set because the API records
          # the generation it resolved.
          gcs_generation = optional(number)

          # Secret Manager secret version holding the script, e.g.
          # "projects/P/secrets/node-init/versions/latest" — for scripts that
          # embed credentials. The node's service account needs
          # secretmanager.versions.access on it.
          secret_manager_secret_uri = optional(string, "")
        }))
      }))

      # Node system-log throughput: DEFAULT (100 KiB/s) or MAX_THROUGHPUT
      # (10 MiB/s, costs a little node CPU) for pools with log-heavy workloads.
      logging_variant = optional(string, "")

      # Flex-start (Dynamic Workload Scheduler): nodes are requested when
      # needed and run up to 7 days at a discount — the on-demand counterpart
      # to queued provisioning for hard-to-get GPU capacity. Immutable.
      flex_start = optional(bool, false)

      # Maximum runtime of each node, in seconds format (e.g. "3600s"), after
      # which it is drained and deleted. Pairs with flex-start/Spot batch
      # pools; leave empty for long-lived pools. Immutable.
      max_run_duration = optional(string, "")

      # Confidential storage on the node disks (hardware-encrypted at the
      # storage layer; pairs with confidential_nodes for end-to-end
      # confidentiality). Requires a supporting machine family and hyperdisk
      # boot disks. Immutable.
      enable_confidential_storage = optional(bool, false)

      # Encryption mode for local SSDs: STANDARD_ENCRYPTION (Google-managed)
      # or EPHEMERAL_KEY_ENCRYPTION (per-node ephemeral keys destroyed with
      # the node — local data is unrecoverable after preemption/deletion).
      # Immutable.
      local_ssd_encryption_mode = optional(string, "")

      # GPUDirect strategy for multi-GPU/multi-node communication (e.g.
      # "GPUDIRECT_TCPX", "GPUDIRECT_TCPXO", "GPUDIRECT_RDMA"). GKE validates
      # the strategy against the machine family and GPU type at apply; the
      # provider passes the value through case-insensitively.
      gpudirect_strategy = optional(string, "")

      # Sole-tenant node group to schedule this pool's nodes onto (nodes on
      # dedicated physical servers you already provisioned). Prefer
      # sole_tenant_config's affinity form for flexible matching. Immutable.
      node_group = optional(string, "")

      # Hyperdisk storage pools the node boot disks are provisioned in —
      # pre-purchased pooled capacity/IOPS shared across disks. Full resource
      # paths. Immutable.
      storage_pools = optional(list(string), [])

      # Resource Manager tags bound to the node VMs (org-policy/firewall
      # tags, distinct from network tags and labels), as
      # {"tagKeys/123": "tagValues/456"} pairs. Mutable.
      resource_manager_tags = optional(map(string), {})

      # Advanced machine features: SMT control, nested virtualization, and
      # the performance monitoring unit. Immutable.
      advanced_machine_features = optional(object({
        # Threads per physical core: 1 disables SMT (licensing or
        # side-channel isolation), 2 keeps it on. Only 1 or 2 are meaningful.
        threads_per_core = optional(number, 0)

        # Nested virtualization on the nodes (running VMs inside pods, e.g.
        # KubeVirt). Requires Haswell-or-newer non-shared-core machine types.
        enable_nested_virtualization = optional(bool)

        # Performance monitoring unit exposure: ARCHITECTURAL (basic counters),
        # STANDARD (most counters), or ENHANCED (all counters) — for profiling
        # workloads that read hardware counters.
        performance_monitoring_unit = optional(string, "")
      }))

      # Boot disk shape as a first-class block — required for hyperdisk
      # tuning (provisioned IOPS/throughput). When set, its disk_type/size
      # take precedence over the flat disk_type/disk_size_gb fields.
      boot_disk = optional(object({
        # Boot disk type (pd-standard, pd-balanced, pd-ssd,
        # hyperdisk-balanced...). Hyperdisk types unlock provisioned_iops /
        # provisioned_throughput below.
        disk_type = optional(string, "")

        # Boot disk size in GB (min 10).
        size_gb = optional(number)

        # Provisioned IOPS — hyperdisk types that support IOPS provisioning
        # only.
        provisioned_iops = optional(number)

        # Provisioned throughput in MiB/s — hyperdisk types that support
        # throughput provisioning only.
        provisioned_throughput = optional(number)
      }))

      # Custom node OS image family, overriding the stock image_type image.
      # Both fields must point at an image built for GKE nodes. Immutable.
      node_image = optional(object({
        # Image family or full image name to boot nodes from.
        image = string

        # Project hosting the image. Empty means the pool's own project.
        image_project = optional(string, "")
      }))

      # Sole-tenant scheduling: affinity rules selecting the sole-tenant node
      # groups this pool's nodes run on, plus a dedicated-vCPU floor.
      # Immutable.
      sole_tenant_config = optional(object({
        # Affinity rules matching sole-tenant node groups. At least one rule
        # selects the groups (e.g. key "compute.googleapis.com/node-group-name",
        # operator IN, values [group name]).
        node_affinities = list(object({
          # Affinity label key, e.g. "compute.googleapis.com/node-group-name".
          key = string

          # IN selects hosts matching values; NOT_IN avoids them.
          operator = string

          # Values matched against the key.
          values = list(string)
        }))

        # Minimum dedicated vCPUs per node reserved for this pool's workloads
        # on the shared physical host.
        min_node_cpus = optional(number)
      }))

      # GKE Sandbox: runs every pod in this pool inside gVisor (user-space
      # kernel isolation) — for untrusted/multi-tenant workloads. The only
      # value is "GVISOR". GKE taints sandbox pools automatically. Immutable.
      sandbox_type = optional(string, "")

      # Windows Server version for WINDOWS_LTSC_CONTAINERD pools:
      # OS_VERSION_LTSC2019 or OS_VERSION_LTSC2022. Immutable.
      windows_os_version = optional(string, "")

      # Host maintenance cadence for the underlying physical hosts:
      # AS_NEEDED (default) or PERIODIC (predictable windows — required by
      # some GPU/TPU shapes). Immutable.
      host_maintenance_interval = optional(string, "")

      # How GKE taints nodes based on CPU architecture: ARM applies the
      # kubernetes.io/arch taint to Arm nodes (the default protection that
      # keeps amd64-only workloads off T2A/Axion pools); NONE disables that
      # automatic taint.
      architecture_taint_behavior = optional(string, "")

      # containerd runtime configuration: private registry access (custom CA
      # domains), per-registry host overrides (mirrors, auth, headers), and
      # writable cgroups.
      containerd_config = optional(object({
        # Trust custom certificate authorities for specific registry domains —
        # required for private registries with self-signed/internal CAs.
        private_registry_access = optional(object({
          # Master toggle for private registry access configuration.
          enabled = optional(bool, false)

          # Per-domain CA trust: each entry names registry FQDNs and the Secret
          # Manager secret holding the CA certificate for them.
          certificate_authority_domains = optional(list(object({
            # Registry FQDNs this CA vouches for, e.g. ["registry.internal:5000"].
            fqdns = list(string)

            # Secret Manager secret URI holding the CA certificate, in the form
            # "projects/{project}/secrets/{secret}/versions/{version}".
            gcp_secret_manager_certificate_uri = string
          })), [])
        }))

        # Per-registry host overrides: mirrors, capabilities, dial timeouts,
        # client certificates, and custom headers, in containerd hosts.toml
        # semantics.
        registry_hosts = optional(list(object({
          # The registry server the overrides apply to, e.g. "docker.io" or
          # "registry.internal:5000".
          server = string

          # Host endpoints serving this registry (mirrors first, in order).
          hosts = optional(list(object({
            # Endpoint URL, e.g. "https://mirror.internal".
            host = string

            # Operations this endpoint can serve — typically "pull" and
            # "resolve" (containerd hosts.toml capability names). GKE validates
            # the values at apply.
            capabilities = optional(list(string), [])

            # Dial timeout for this endpoint, e.g. "10s".
            dial_timeout = optional(string, "")

            # Path override on the endpoint host (registry served under a
            # non-standard path).
            override_path = optional(bool)

            # Secret Manager secret URI for the CA certificate to trust for this
            # endpoint.
            ca_secret_uri = optional(string, "")

            # Secret Manager secret URI for the client TLS certificate presented to
            # this endpoint.
            client_cert_secret_uri = optional(string, "")

            # Secret Manager secret URI for the client TLS key presented to this
            # endpoint.
            client_key_secret_uri = optional(string, "")

            # Custom HTTP headers sent to this endpoint, e.g. authorization
            # headers for pull-through caches.
            headers = optional(map(string), {})
          })), [])
        })), [])

        # Writable cgroup filesystem inside containers — for workloads that
        # manage their own sub-cgroups (nested container runtimes, some JVMs).
        writable_cgroups_enabled = optional(bool)
      }))
    }))

    # Destroy-time stance of the IaC engines toward the pool itself:
    # DELETE (default) destroys it, PREVENT fails any plan that would
    # destroy it, ABANDON removes it from state and leaves the pool
    # running in GCP. This is an engine-side control, not a GKE API field.
    deletion_policy = optional(string, "")

    # Skips the per-pool Instance Group Manager queries that reconcile the
    # observed node count on every plan — a quota/performance optimization
    # for very large pools. While true, node-count drift is invisible to
    # plans and the instance_group_urls outputs go stale. The engines
    # already never fight the autoscaler over node_count; this additionally
    # silences the read-side queries.
    ignore_node_count_changes = optional(bool, false)

    # How nodes drain when the POOL ITSELF is deleted or replaced: grace
    # periods and whether PodDisruptionBudgets are honored during the
    # teardown. Distinct from upgrade_settings, which paces upgrades of a
    # pool that continues to exist. NOTE: customized node drain requires
    # project-level enablement from GCP support — on a project without it,
    # the API rejects the create with "customized node drain timeout is not
    # enabled for this project, please contact your account manager or open
    # a support case to enable it".
    node_drain_config = optional(object({
      # Grace period each node gets to finish draining before it is removed,
      # seconds format, e.g. "300s".
      grace_termination_duration = optional(string, "")

      # How long the drain waits on a blocking PodDisruptionBudget before
      # proceeding anyway, seconds format, e.g. "3600s".
      pdb_timeout_duration = optional(string, "")

      # Honor PodDisruptionBudgets while the pool is being deleted (bounded
      # by pdb_timeout_duration) instead of evicting immediately.
      respect_pdb_during_node_pool_deletion = optional(bool)
    }))

    # Holds this pool on its current Kubernetes version until that
    # version's end-of-support date, exempting it from GKE's automatic
    # upgrades (the cluster's maintenance windows and exclusions still
    # govern everything else). For a workload that must not move minor
    # versions until it has been re-qualified. GKE reports the resulting
    # exclusion window (start and end) in the node pool's status; when the
    # version reaches end of support the exclusion lapses and upgrades
    # resume.
    exclude_upgrades_until_end_of_support = optional(bool, false)
  })
}
