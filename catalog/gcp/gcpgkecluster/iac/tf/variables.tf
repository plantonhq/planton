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
  description = "GcpGkeCluster specification"
  type = object({
    # The GCP project in which to create this cluster.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Example: "my-prod-project-123"
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the GKE cluster in GCP. Immutable. If not specified, defaults to
    # metadata.name. Must be 1-40 characters: lowercase letters, digits, and
    # hyphens; starting with a letter and ending with a letter or digit.
    # Example: "prod-primary"
    cluster_name = optional(string, "")

    # Location of the cluster's control plane. Immutable. A region
    # ("us-central1") creates a REGIONAL cluster — control-plane replicas in
    # three zones, higher availability, and node_locations defaulting to all
    # zones in the region. A zone ("us-central1-a") creates a ZONAL cluster —
    # a single control-plane instance, cheaper, with a brief control-plane
    # outage during upgrades. Production clusters should be regional.
    location = string

    # Human-readable description of the cluster. Immutable.
    description = optional(string, "")

    # The VPC network the cluster lives in. Accepts a network self link or a
    # reference to a GcpVpcNetwork resource. Immutable. An explicit network is
    # deliberately required — clusters on the legacy auto-created "default"
    # network do not compose into reviewable infrastructure.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    network = string

    # The subnetwork nodes are attached to. Accepts a subnetwork self link or
    # a reference to a GcpSubnetwork resource. Immutable. Must be in the same
    # region as the cluster location. Pod and service ranges come from this
    # subnetwork's secondary ranges (see ip_allocation).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnetwork = string

    # Zones in which nodes (not the control plane) run, e.g.
    # ["us-central1-a", "us-central1-b"]. For a regional cluster this narrows
    # node placement from all zones in the region; for a zonal cluster it adds
    # node zones beyond the control-plane zone (multi-zonal). Mutable.
    node_locations = optional(list(string), [])

    # GCE resource labels applied to the cluster (merged with the standard
    # platform labels). Mutable. These are cloud-billing/inventory labels, not
    # Kubernetes object labels.
    resource_labels = optional(map(string), {})

    # Engine-side delete guard: while true (the default, matching GCP), both
    # IaC engines refuse to destroy the cluster — the plan/preview fails until
    # this is set to false. A cluster deletion destroys every workload on it;
    # keep this on for anything real.
    deletion_protection = optional(bool)

    # VPC-native pod/service IP allocation. If omitted, GKE creates and
    # manages the secondary ranges itself — fine for dev clusters; production
    # clusters should name planned ranges on the subnetwork for address-space
    # governance.
    ip_allocation = optional(object({
      # Name of an existing secondary range on the subnetwork for POD IPs.
      # Accepts a literal range name or a reference to a GcpSubnetwork
      # resource's secondary ranges. Immutable.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      cluster_secondary_range_name = optional(string, "")

      # Name of an existing secondary range on the subnetwork for SERVICE
      # (ClusterIP) IPs. Immutable.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      services_secondary_range_name = optional(string, "")

      # CIDR block for a GKE-created pod range (e.g. "10.4.0.0/14"), or a
      # netmask size (e.g. "/14") to let GKE pick the space. Immutable.
      cluster_ipv4_cidr_block = optional(string, "")

      # CIDR block for a GKE-created services range (e.g. "10.8.0.0/20"), or a
      # netmask size. Immutable.
      services_ipv4_cidr_block = optional(string, "")

      # IP stack of the cluster: IPV4 (default) or IPV4_IPV6 (dual-stack;
      # requires a dual-stack subnetwork). Immutable.
      stack_type = optional(string)

      # Names of ADDITIONAL existing secondary ranges node pools may use for
      # pod IPs — the mechanism for growing pod address space after creation
      # (added ranges are mutable; the primary pod range is not).
      additional_pod_range_names = optional(list(string), [])

      # Disables the 2x pod-CIDR overprovisioning GKE applies per node by
      # default — doubles node density per pod range at the cost of headroom
      # for pod churn. Immutable.
      pod_cidr_overprovision_disabled = optional(bool, false)

      # Additional SUBNETWORKS whose secondary ranges node pools may draw pod
      # IPs from — grows pod address space across subnetworks (beyond
      # additional_pod_range_names, which only names more ranges on the
      # cluster's own subnetwork).
      additional_ip_ranges = optional(list(object({
        # The subnetwork carrying the ranges. Accepts a self link or a
        # reference to a GcpSubnetwork resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        subnetwork = string

        # Secondary range names on that subnetwork usable for pod IPs.
        pod_ipv4_range_names = optional(list(string), [])

        # Lifecycle status of the subnetwork for scheduling: set "DRAINING" to
        # stop NEW node pools from selecting it (existing pools keep running).
        status = optional(string, "")
      })), [])

      # Automatic IP address management: GKE plans and allocates the pod and
      # service ranges itself (no manual CIDR planning). Immutable.
      auto_ipam_enabled = optional(bool, false)

      # Network tier for the cluster's IP allocation (e.g. "PREMIUM" or
      # "STANDARD"). GKE validates accepted tiers at apply.
      network_tier = optional(string, "")
    }))

    # The cluster dataplane. ADVANCED_DATAPATH is Dataplane V2 (eBPF/Cilium):
    # built-in NetworkPolicy enforcement without Calico, dataplane
    # observability, and FQDN/Cilium policy support — the recommended choice
    # for new clusters. LEGACY_DATAPATH is kube-proxy/iptables. Immutable.
    datapath_provider = optional(string, "")

    # Default maximum pods per node for the cluster (8-256, default 110).
    # Lower values shrink the per-node pod CIDR slice and stretch the pod
    # range across more nodes. Immutable; node pools can override it.
    default_max_pods_per_node = optional(number)

    # Mirrors pod-to-pod traffic on the same node to the VPC dataplane, making
    # it visible to VPC flow logs and packet mirroring. Mutable.
    enable_intranode_visibility = optional(bool, false)

    # GKE subsetting for internal L4 load balancers: backends become a subset
    # of nodes instead of all nodes, lifting the 250-node ILB ceiling for
    # large clusters. Enabling is one-way — it cannot be turned off without
    # recreating the cluster.
    enable_l4_ilb_subsetting = optional(bool, false)

    # Allows FQDN (domain-name based) NetworkPolicy rules. Requires Dataplane
    # V2 (datapath_provider ADVANCED_DATAPATH). Mutable.
    enable_fqdn_network_policy = optional(bool, false)

    # Allows CiliumClusterwideNetworkPolicy objects (cluster-scoped L3/L4
    # policy). Requires Dataplane V2. Mutable.
    enable_cilium_clusterwide_network_policy = optional(bool, false)

    # Enables multi-networking: pods can attach to additional node network
    # interfaces (Network/GKENetworkParamSet objects). Immutable; requires
    # Dataplane V2.
    enable_multi_networking = optional(bool, false)

    # Outbound-only private IPv6 access from nodes/pods to Google services
    # (PRIVATE_IPV6_GOOGLE_ACCESS_TO_GOOGLE) or bidirectional
    # (PRIVATE_IPV6_GOOGLE_ACCESS_BIDIRECTIONAL); DISABLED turns it off.
    private_ipv6_google_access = optional(string, "")

    # Encrypts inter-node pod traffic transparently (Dataplane V2 only):
    # INTER_NODE_TRANSPARENT encrypts, IN_TRANSIT_ENCRYPTION_DISABLED does
    # not. Zero application changes; small latency cost.
    in_transit_encryption = optional(string, "")

    # Disables the default source NAT for pod IPs leaving the cluster —
    # required for some routable-pod (non-masquerade) network designs where
    # pod IPs must be preserved end to end. Mutable.
    disable_default_snat = optional(bool, false)

    # Enables Kubernetes NetworkPolicy enforcement via Calico (the legacy
    # enforcement path, paired with the network-policy addon). On Dataplane V2
    # clusters leave this off — NetworkPolicy is enforced natively. Mutable.
    enable_network_policy = optional(bool, false)

    # Cluster DNS provider configuration (Cloud DNS instead of kube-dns).
    # If omitted, GKE uses its platform default.
    dns_config = optional(object({
      # In-cluster DNS provider: CLOUD_DNS (managed, no kube-dns pods to
      # scale), KUBE_DNS (explicit kube-dns), or PLATFORM_DEFAULT (GKE
      # chooses).
      cluster_dns = optional(string, "")

      # Cloud DNS scope: CLUSTER_SCOPE (records visible in-cluster only) or
      # VPC_SCOPE (cluster DNS records resolvable across the whole VPC —
      # enables VMs to resolve headless services etc.).
      cluster_dns_scope = optional(string, "")

      # Custom cluster DNS suffix (default "cluster.local").
      cluster_dns_domain = optional(string, "")

      # With CLUSTER_SCOPE Cloud DNS, ALSO publishes services under this domain
      # VPC-wide — cluster-scoped DNS plus selective VPC visibility.
      additive_vpc_scope_dns_domain = optional(string, "")
    }))

    # Gateway API support: CHANNEL_STANDARD installs the Gateway API CRDs
    # and the GKE Gateway controller (the successor to Ingress);
    # CHANNEL_EXPERIMENTAL adds experimental-channel CRDs;
    # CHANNEL_DISABLED turns it off. Mutable.
    gateway_api_channel = optional(string, "")

    # Allows Services of type LoadBalancer/ClusterIP to use external IPs.
    # Off by default as a security posture (CVE-2020-8554 mitigation) —
    # enable only if a workload genuinely needs external IPs on Services.
    enable_service_external_ips = optional(bool, false)

    # Total egress bandwidth tier for node network performance. TIER_1
    # unlocks up to 100 Gbps on supported machine families (N2/N2D/C2/C3...);
    # requires gVNIC on the node pools that use it.
    total_egress_bandwidth_tier = optional(string, "")

    # Stops GKE from reconciling the VPC firewall rules it creates for L4
    # load balancers — for environments where firewall rules are governed
    # externally and GKE's reconciliation fights the desired state.
    disable_l4_lb_firewall_reconciliation = optional(bool, false)

    # Private-cluster topology: private nodes (no public node IPs) and/or a
    # private-only control-plane endpoint. If omitted, nodes get public IPs
    # and the control plane has a public endpoint — fine for sandboxes, not
    # for production.
    private_cluster = optional(object({
      # Nodes get only internal IPs. Outbound internet (image pulls from
      # non-Google registries, external APIs) then requires Cloud NAT on the
      # network — compose a GcpRouterNat.
      enable_private_nodes = optional(bool, false)

      # Removes the control plane's PUBLIC endpoint entirely: kubectl works
      # only from inside the VPC (or via the DNS endpoint / authorized
      # networks' private enforcement). The strictest posture; requires
      # enable_private_nodes.
      enable_private_endpoint = optional(bool, false)

      # RFC1918 /28 block for the control plane on peering-based private
      # clusters, e.g. "172.16.0.16/28". Must not overlap any range in the VPC.
      # Immutable. Newer PSC-based clusters don't need it.
      master_ipv4_cidr_block = optional(string, "")

      # Subnetwork in which the control plane's private endpoint is placed
      # (PSC-based private clusters). Accepts a self link or a GcpSubnetwork
      # reference. Immutable.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      private_endpoint_subnetwork = optional(string, "")

      # Makes the private control-plane endpoint reachable from other GCP
      # regions and on-prem over interconnect/VPN (not just the cluster's
      # region).
      enable_master_global_access = optional(bool, false)
    }))

    # CIDR allowlist for control-plane API access. Without it, the public
    # endpoint (when enabled) accepts connections from any IP that presents
    # valid credentials.
    master_authorized_networks = optional(object({
      # CIDR ranges allowed to reach the control-plane endpoint(s). An empty
      # list with the block present means "no external networks authorized".
      cidr_blocks = optional(list(object({
        # The CIDR range, e.g. "203.0.113.0/24".
        cidr_block = string

        # Display label for this entry in the console.
        display_name = optional(string, "")
      })), [])

      # Whether Google Cloud public IPs (e.g. Cloud Shell, Cloud Build default
      # pools) may reach the control plane.
      gcp_public_cidrs_access_enabled = optional(bool)

      # Enforces the allowlist on the PRIVATE endpoint too (by default only the
      # public endpoint is filtered).
      private_endpoint_enforcement_enabled = optional(bool)
    }))

    # Control-plane endpoint surface: the DNS endpoint (a Google-managed
    # *.gke.goog name reachable without VPC peering — the modern access path)
    # and the IP endpoints toggle.
    control_plane_endpoints = optional(object({
      # Allows the Google-managed DNS endpoint (*.gke.goog) to accept traffic
      # from outside the cluster's VPC — IAM-authenticated access without
      # peering, bastions, or public IPs. The modern alternative to juggling
      # authorized networks.
      dns_endpoint_allow_external_traffic = optional(bool, false)

      # Whether IP-based endpoints (the classic public/private IPs) are served
      # at all. Set false for DNS-endpoint-only clusters — the strongest
      # endpoint posture.
      ip_endpoints_enabled = optional(bool)

      # Serve Kubernetes ServiceAccount TOKENS via the DNS endpoint —
      # workloads outside the VPC can authenticate without IP connectivity to
      # the control plane.
      enable_k8s_tokens_via_dns = optional(bool)

      # Serve Kubernetes client CERTIFICATES via the DNS endpoint.
      enable_k8s_certs_via_dns = optional(bool)
    }))

    # Kubernetes release channel for automatic upgrades. REGULAR (default)
    # balances freshness and stability; RAPID gets new minors first; STABLE
    # lags for maximum soak; EXTENDED keeps a minor supported longer for slow
    # movers; NONE opts out of channel-based auto-upgrade (you own version
    # management via min_master_version).
    release_channel = optional(string)

    # Minimum Kubernetes version for the control plane, e.g. "1.31" or
    # "1.31.4-gke.1256000". GKE may run a newer patch. Use with release
    # channel NONE for pinned-version clusters; on a channel, prefer letting
    # the channel drive versions.
    min_master_version = optional(string, "")

    # Maintenance windows and exclusions controlling WHEN GKE may perform
    # automatic control-plane and node maintenance.
    maintenance_policy = optional(object({
      # Same 4-hour window every day, starting at this UTC time.
      daily_window = optional(object({
        # Start of the daily 4-hour window, "HH:MM" (UTC), e.g. "03:00".
        start_time = string
      }))

      # RRULE-based recurring window anchored on an absolute first
      # occurrence (start_time/end_time as RFC3339 timestamps) — e.g.
      # weekends only. Finer control than the daily window.
      recurring_window = optional(object({
        # Window start, RFC3339, e.g. "2026-01-01T02:00:00Z".
        start_time = string

        # Window end (defines the window LENGTH; recurrence drives repetition),
        # RFC3339.
        end_time = string

        # RFC5545 RRULE, e.g. "FREQ=WEEKLY;BYDAY=SA,SU" for weekends.
        recurrence = string
      }))

      # RRULE-based recurring window expressed as a time of day plus a
      # duration, with an optional date the recurrence may first start. The
      # same recurrence power as recurring_window without committing to one
      # absolute first timestamp — pick this when the policy is authored as
      # "every Saturday at 02:00 for 6 hours, starting next quarter".
      recurring_time_window = optional(object({
        # Time of day (UTC) each window instance begins.
        window_start_time = object({
          # Hour of the day, 0-23.
          hours = optional(number, 0)

          # Minute of the hour, 0-59.
          minutes = optional(number, 0)

          # Second of the minute, 0-59.
          seconds = optional(number, 0)
        })

        # Length of each window instance as a duration string with a unit
        # suffix, e.g. "4h", "6h30m", "21600s". Must be positive.
        window_duration = string

        # RFC5545 RRULE, e.g. "FREQ=WEEKLY;BYDAY=SA,SU" for weekends.
        recurrence = string

        # Earliest calendar date the recurrence may start; window instances
        # before it are skipped. Leave unset to start immediately.
        delay_until = optional(object({
          # Four-digit year, e.g. 2027.
          year = optional(number, 0)

          # Month of the year, 1-12.
          month = optional(number, 0)

          # Day of the month, 1-31.
          day = optional(number, 0)
        }))
      }))

      # Date ranges during which non-emergency maintenance is blocked (change
      # freezes). Maximum 20; the allowed scope/duration depends on the release
      # channel.
      exclusions = optional(list(object({
        # Name of the exclusion, e.g. "year-end-freeze".
        exclusion_name = string

        # Exclusion start, RFC3339.
        start_time = string

        # Exclusion end, RFC3339.
        end_time = string

        # What the exclusion blocks: NO_UPGRADES (everything),
        # NO_MINOR_UPGRADES, or NO_MINOR_OR_NODE_UPGRADES. Empty means
        # NO_UPGRADES.
        scope = optional(string, "")

        # UNTIL_END_OF_SUPPORT stretches the exclusion until the running minor
        # version's end of support, instead of stopping at end_time — the
        # "never force this cluster off its minor" stance. Requires scope.
        end_time_behavior = optional(string, "")
      })), [])

      # Minimum spacing between consecutive disruptive maintenance events —
      # a floor on how often GKE may disrupt the cluster with version
      # changes, independent of windows.
      disruption_budget = optional(object({
        # Minimum interval between MINOR version disruptions, seconds format,
        # e.g. "2419200s" (28 days).
        minor_version_disruption_interval = optional(string, "")

        # Minimum interval between PATCH version disruptions, seconds format.
        patch_version_disruption_interval = optional(string, "")
      }))
    }))

    # Node auto-provisioning (NAP): GKE creates and deletes node pools
    # automatically within the resource limits you set — the cluster-level
    # autoscaler above individual node-pool autoscaling.
    cluster_autoscaling = optional(object({
      # Whether node auto-provisioning is on.
      enabled = optional(bool, false)

      # Cluster-wide bounds per resource type ("cpu", "memory", or an
      # accelerator like "nvidia-tesla-t4"). NAP never provisions beyond the
      # maximums — the cost brake.
      resource_limits = optional(list(object({
        # "cpu", "memory", or an accelerator type such as "nvidia-tesla-t4".
        resource_type = string

        # Minimum amount kept provisioned (0 = none).
        minimum = optional(number, 0)

        # Maximum amount NAP may provision. Required — an unbounded NAP is an
        # unbounded bill.
        maximum = optional(number, 0)
      })), [])

      # BALANCED (default) or OPTIMIZE_UTILIZATION (scales down harder and
      # faster — denser packing, more pod churn).
      autoscaling_profile = optional(string)

      # Zones in which NAP may create node pools (defaults to the cluster's
      # node zones).
      auto_provisioning_locations = optional(list(string), [])

      # Defaults applied to every node pool NAP creates (identity, disks,
      # image, shielding, auto-repair/upgrade).
      auto_provisioning_defaults = optional(object({
        # IAM service account NAP-created nodes run as. Accepts an email or a
        # reference to a GcpServiceAccount resource. Defaults to the Compute
        # Engine default SA — create a minimal dedicated SA for production.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        service_account = optional(string, "")

        # OAuth scopes on NAP-created nodes. With Workload Identity, the
        # cloud-platform scope is the norm (IAM governs actual access).
        oauth_scopes = optional(list(string), [])

        # Boot disk size in GB for NAP-created nodes (default 100).
        disk_size_gb = optional(number)

        # Boot disk type: pd-standard (default), pd-balanced, pd-ssd, or
        # hyperdisk-balanced.
        disk_type = optional(string, "")

        # Node image, e.g. "COS_CONTAINERD" (default) or "UBUNTU_CONTAINERD"
        # (legacy "COS"/"UBUNTU" accepted for old clusters).
        image_type = optional(string, "")

        # Minimum CPU platform, e.g. "Intel Ice Lake".
        min_cpu_platform = optional(string, "")

        # Customer-managed key encrypting NAP-created node boot disks. Accepts a
        # full crypto key path or a reference to a GcpKmsKey resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        boot_disk_kms_key = optional(string, "")

        # Shielded-VM secure boot on NAP-created nodes (GCP default false).
        enable_secure_boot = optional(bool, false)

        # Shielded-VM integrity monitoring on NAP-created nodes (GCP default
        # true).
        enable_integrity_monitoring = optional(bool)

        # Automatic node upgrades on NAP-created pools (GCP default true;
        # required on a release channel).
        auto_upgrade = optional(bool)

        # Automatic node repair on NAP-created pools (GCP default true).
        auto_repair = optional(bool)

        # How upgrades roll through NAP-created pools: surge settings or
        # blue-green, in the same shape as a GcpGkeNodePool's upgrade_settings.
        upgrade_settings = optional(object({
          # Additional nodes added during a surge upgrade.
          max_surge = optional(number)

          # Nodes that may be simultaneously unavailable during a surge upgrade.
          max_unavailable = optional(number)

          # SURGE (rolling replacement) or BLUE_GREEN (full new node set,
          # migrate, soak, delete).
          strategy = optional(string, "")

          # Blue-green rollout pacing. Only with strategy BLUE_GREEN.
          blue_green_settings = optional(object({
            # How the old (blue) node set drains, batch by batch.
            standard_rollout_policy = optional(object({
              # Fraction of blue nodes drained per batch, 0.0-1.0.
              batch_percentage = optional(number)

              # Number of blue nodes drained per batch.
              batch_node_count = optional(number)

              # Soak time after each batch, seconds format, e.g. "600s".
              batch_soak_duration = optional(string, "")
            }))

            # Soak time after the blue set is fully drained before deletion,
            # seconds format, e.g. "3600s".
            node_pool_soak_duration = optional(string, "")
          }))
        }))
      }))

      # Default compute classes: NAP provisions through the cluster's default
      # ComputeClass definitions instead of the legacy per-resource limits
      # path — the newer Autopilot-style provisioning model on Standard
      # clusters.
      default_compute_class_enabled = optional(bool)
    }))

    # Enables Vertical Pod Autoscaling: recommends (and can apply) per-pod
    # CPU/memory requests based on observed usage. Mutable.
    enable_vertical_pod_autoscaling = optional(bool, false)

    # Horizontal Pod Autoscaler profile. PERFORMANCE makes HPA react faster
    # (higher-resolution metrics pipeline); NONE is the classic behavior.
    hpa_profile = optional(string, "")

    # Workload Identity Federation for GKE: pods authenticate to GCP APIs as
    # IAM service accounts via the cluster's workload pool
    # (PROJECT_ID.svc.id.goog) — no exported service-account keys. Enabled by
    # default; disabling it forces workloads back to node service-account
    # scopes or key files, which is almost never the right call.
    workload_identity_enabled = optional(bool)

    # Shielded GKE nodes: secure boot + integrity monitoring on node VMs,
    # verifying node provenance cryptographically. GCP's default (true);
    # leave unset on Autopilot clusters (always shielded).
    enable_shielded_nodes = optional(bool)

    # Envelope encryption of Kubernetes secrets at the application layer with
    # a Cloud KMS key (CMEK for etcd secrets).
    database_encryption = optional(object({
      # ENCRYPTED (Kubernetes secrets wrapped with key_name),
      # ALL_OBJECTS_ENCRYPTION_ENABLED (every etcd object wrapped, not just
      # secrets), or DECRYPTED.
      state = string

      # The Cloud KMS crypto key (full path) used for envelope encryption.
      # Accepts a literal or a reference to a GcpKmsKey resource. The GKE
      # service agent needs Encrypter/Decrypter on it.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      key_name = optional(string, "")
    }))

    # Binary Authorization: PROJECT_SINGLETON_POLICY_ENFORCE admits only
    # container images that satisfy the project's Binary Authorization policy
    # (signature/attestation-based supply-chain control); DISABLED turns
    # enforcement off.
    binary_authorization_evaluation_mode = optional(string, "")

    # GKE Security Posture dashboard: workload configuration auditing and
    # vulnerability scanning surfaced in the console.
    security_posture = optional(object({
      # Workload configuration auditing: DISABLED, BASIC, or ENTERPRISE.
      mode = optional(string, "")

      # Workload vulnerability scanning: VULNERABILITY_DISABLED,
      # VULNERABILITY_BASIC, or VULNERABILITY_ENTERPRISE.
      vulnerability_mode = optional(string, "")
    }))

    # Google Group for RBAC: members of this group (and its subgroups) can be
    # referenced in RBAC bindings. Must be named gke-security-groups@YOURDOMAIN.
    authenticator_security_group = optional(string, "")

    # Legacy Attribute-Based Access Control. Leave off: ABAC predates RBAC
    # and grants coarse permissions that defeat modern authorization. Exists
    # only for very old workloads being migrated.
    enable_legacy_abac = optional(bool, false)

    # Issues workload mTLS certificates via the mesh CA (used by Anthos
    # Service Mesh / managed Istio). Requires Workload Identity.
    enable_mesh_certificates = optional(bool, false)

    # Built-in Secret Manager add-on: mounts Secret Manager secrets into pods
    # via the CSI driver without third-party operators.
    enable_secret_manager_csi = optional(bool, false)

    # Confidential GKE nodes: node VMs run with hardware memory encryption
    # (AMD SEV / SEV-SNP or Intel TDX). Immutable; restricts machine families.
    confidential_nodes = optional(object({
      # Whether Confidential GKE Nodes are enabled cluster-wide. Immutable.
      enabled = optional(bool, false)

      # Confidential computing technology: SEV (default), SEV_SNP, or TDX.
      # Machine-family support varies by choice.
      confidential_instance_type = optional(string, "")
    }))

    # Anonymous Kubernetes API authentication posture. LIMITED restricts the
    # anonymous user to the health-check endpoints only (the hardened
    # default on new versions); ENABLED preserves full legacy anonymous
    # access subject to RBAC.
    anonymous_authentication_mode = optional(string, "")

    # GKE Identity Service: authenticate to the Kubernetes API with external
    # OIDC identity providers (beyond Google accounts).
    enable_identity_service = optional(bool, false)

    # Which cluster components ship logs to Cloud Logging. If omitted, GKE's
    # default (system components + workloads) applies; declare it with
    # `enabled: false` to turn the integration off.
    logging = optional(object({
      # Components exposing logs: SYSTEM_COMPONENTS, WORKLOADS, APISERVER,
      # CONTROLLER_MANAGER, SCHEDULER, KCP_CONNECTION, KCP_SSHD, KCP_HPA,
      # KCP_VPA.
      components = optional(list(string), [])

      # Whether Cloud Logging integration is on. Unset means on: declaring the
      # block with components has always meant shipping their logs, and
      # `enabled: false` is the explicit OFF (GKE then receives an empty
      # component list).
      enabled = optional(bool)
    }))

    # Which cluster components ship metrics to Cloud Monitoring, plus managed
    # Prometheus. If omitted, GKE's defaults apply (system metrics + managed
    # Prometheus on current versions).
    monitoring = optional(object({
      # Components exposing metrics: SYSTEM_COMPONENTS, APISERVER, SCHEDULER,
      # CONTROLLER_MANAGER, STORAGE, HPA, POD, DAEMONSET, DEPLOYMENT,
      # STATEFULSET, KUBELET, CADVISOR, DCGM, JOBSET. An empty list disables
      # Cloud Monitoring integration.
      components = optional(list(string), [])

      # Google Cloud Managed Service for Prometheus: managed collection of
      # Prometheus metrics (GKE's default on current versions). Disabling it
      # means running your own Prometheus stack.
      managed_prometheus_enabled = optional(bool)

      # Managed Prometheus auto-monitoring scope: ALL deploys packaged
      # PodMonitorings for supported workloads automatically; NONE leaves
      # monitoring configuration entirely to you.
      auto_monitoring_scope = optional(string, "")

      # Dataplane V2 observability metrics (per-flow network telemetry).
      advanced_datapath_metrics_enabled = optional(bool, false)

      # Dataplane V2 observability relay (Hubble-compatible flow export for
      # in-cluster observability tooling).
      advanced_datapath_relay_enabled = optional(bool, false)
    }))

    # Cluster lifecycle notifications (upgrades, security bulletins)
    # published to a Pub/Sub topic — the hook for upgrade automation and
    # fleet dashboards.
    notification_pubsub = optional(object({
      # Whether notifications are published.
      enabled = optional(bool, false)

      # The Pub/Sub topic (projects/{project}/topics/{name}). Accepts a literal
      # or a reference to a GcpPubSubTopic resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      topic = optional(string, "")

      # Restrict to specific event types; empty publishes all. Values:
      # UPGRADE_EVENT, UPGRADE_AVAILABLE_EVENT, SECURITY_BULLETIN_EVENT,
      # UPGRADE_INFO_EVENT.
      event_types = optional(list(string), [])
    }))

    # Cost allocation: exports per-namespace/per-label resource usage into
    # the billing export, so cluster cost can be attributed to teams.
    enable_cost_management = optional(bool, false)

    # Continuous resource-usage metering into a BigQuery dataset (network
    # egress and resource consumption records).
    resource_usage_export = optional(object({
      # The BigQuery dataset receiving usage records. Accepts a dataset ID or a
      # reference to a GcpBigQueryDataset resource. The dataset must be in the
      # cluster's project.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      bigquery_dataset_id = string

      # Also meter pod network egress (deploys a metering agent per node).
      enable_network_egress_metering = optional(bool, false)

      # Meter resource requests/consumption (GCP default true).
      enable_resource_consumption_metering = optional(bool)
    }))

    # Cluster addons. If omitted, GKE defaults apply: HTTP load balancing,
    # horizontal pod autoscaling, and the PD CSI driver enabled; everything
    # else disabled.
    addons = optional(object({
      # The GCE ingress controller backing Kubernetes Ingress/Gateway with
      # Google Cloud Load Balancers. Disable only when bringing your own
      # ingress stack end to end.
      http_load_balancing_enabled = optional(bool)

      # The Horizontal Pod Autoscaler controller.
      horizontal_pod_autoscaling_enabled = optional(bool)

      # The Compute Engine Persistent Disk CSI driver (dynamic PD volumes).
      gce_persistent_disk_csi_driver_enabled = optional(bool)

      # The Filestore CSI driver (managed NFS volumes).
      gcp_filestore_csi_driver_enabled = optional(bool, false)

      # The Cloud Storage FUSE CSI driver (mount GCS buckets as volumes).
      gcs_fuse_csi_driver_enabled = optional(bool, false)

      # The Backup for GKE agent (workload + volume backup/restore).
      gke_backup_agent_enabled = optional(bool, false)

      # NodeLocal DNSCache (per-node DNS cache; reduces DNS latency and
      # kube-dns/Cloud DNS load). Enabling recreates node pools on existing
      # clusters.
      dns_cache_enabled = optional(bool, false)

      # Config Connector (manage GCP resources through Kubernetes CRDs).
      config_connector_enabled = optional(bool, false)

      # Stateful HA operator (faster failover for stateful workloads).
      stateful_ha_enabled = optional(bool, false)

      # The Ray operator (KubeRay) for distributed Python/AI workloads.
      ray_operator_enabled = optional(bool, false)

      # Ray cluster logging integration (with ray_operator_enabled).
      ray_cluster_logging_enabled = optional(bool, false)

      # Ray cluster monitoring integration (with ray_operator_enabled).
      ray_cluster_monitoring_enabled = optional(bool, false)

      # Cloud Run for Anthos / Knative serving addon.
      cloudrun_enabled = optional(bool, false)

      # Load balancer type for the Cloud Run addon:
      # LOAD_BALANCER_TYPE_INTERNAL serves it on an internal LB instead of
      # the default external one.
      cloudrun_load_balancer_type = optional(string, "")

      # The Parallelstore CSI driver (managed parallel filesystem for
      # AI/HPC).
      parallelstore_csi_driver_enabled = optional(bool, false)

      # The Lustre CSI driver (Managed Lustre high-performance filesystem).
      lustre_csi_driver_enabled = optional(bool, false)

      # Serve the Lustre protocol on the legacy port (compatibility with
      # pre-GA Lustre deployments; with lustre_csi_driver_enabled).
      lustre_csi_legacy_port_enabled = optional(bool, false)

      # Disable multi-NIC transport for the Lustre CSI driver (with
      # lustre_csi_driver_enabled).
      lustre_csi_disable_multi_nic = optional(bool, false)

      # Pod snapshot support (checkpoint/restore of running pods).
      pod_snapshot_enabled = optional(bool, false)

      # GKE Sandbox (gVisor) agent addon for Autopilot sandbox pods.
      agent_sandbox_enabled = optional(bool, false)

      # The slice controller addon (TPU slice management).
      slice_controller_enabled = optional(bool, false)

      # The Slurm operator addon (Slurm-on-GKE for HPC scheduling).
      slurm_operator_enabled = optional(bool, false)

      # High Scale Checkpointing: the addon that lets large AI/ML training
      # jobs checkpoint and restore state at scale (multi-tier checkpointing
      # onto node-local and Cloud Storage tiers) so a job resumes from its
      # last checkpoint after a preemption or failure instead of restarting.
      high_scale_checkpointing_enabled = optional(bool, false)

      # Node Readiness Controller: the addon that holds a node out of
      # scheduling until its readiness rules (for example, required daemon
      # pods or device drivers) pass, so workloads never land on a node
      # whose accelerators or networking are not yet usable.
      node_readiness_controller_enabled = optional(bool, false)
    }))

    # Creates an Autopilot cluster: GKE provisions and manages nodes, bills
    # per pod, and enforces a hardened posture. Immutable. Autopilot clusters
    # take no GcpGkeNodePool resources and reject the node-management fields
    # guarded by validation rules above.
    enable_autopilot = optional(bool, false)

    # Autopilot only: permit workloads with NET_ADMIN capability (needed by
    # some networking agents/service meshes on Autopilot).
    allow_net_admin = optional(bool, false)

    # Registers the cluster with a fleet in the given project (the hub for
    # multi-cluster features: multi-cluster ingress/services, config
    # management, team scopes).
    fleet_project = optional(string, "")

    # Fleet membership type. LIGHTWEIGHT registers a lightweight membership
    # (reduced fleet feature surface, no Connect agent). Empty uses the
    # fleet default (full membership).
    fleet_membership_type = optional(string, "")

    # Destroy-time stance of the IaC engines toward the cluster itself:
    # DELETE (default) destroys it, PREVENT fails any plan that would
    # destroy it, ABANDON removes it from state and leaves the cluster
    # running in GCP. This is an engine-side control layered UNDER
    # deletion_protection (the GKE-native guard): deletion_protection
    # blocks the API call itself; deletion_policy governs what the engines
    # even attempt.
    deletion_policy = optional(string, "")

    # Skips the per-pool Instance Group Manager queries during cluster
    # reads — a quota/performance optimization for clusters with many
    # pools. While true, node-count drift is invisible to plans and
    # managed-instance-group outputs go stale on every pool.
    ignore_node_count_changes = optional(bool, false)

    # Skips refreshing inline node-pool state from the API during cluster
    # reads — a substantial plan/apply speedup on clusters with many pools.
    # Safe in this composition because node pools are always separate
    # GcpGkeNodePool resources, never inline blocks on the cluster.
    skip_node_pool_refresh = optional(bool, false)

    # Creates an ALPHA cluster: all Kubernetes alpha feature gates enabled,
    # no SLA, cannot be upgraded, and GKE DELETES the cluster after 30
    # days. Strictly for short-lived feature evaluation. Immutable.
    enable_kubernetes_alpha = optional(bool, false)

    # Specific Kubernetes BETA API groups enabled on the cluster, e.g.
    # ["resource.k8s.io/v1beta1/deviceclasses"]. Beta APIs are off by
    # default on new clusters; list exactly the groups a workload needs.
    k8s_beta_apis = optional(list(string), [])

    # Dataplane optimization mode (pass-through to the GKE API; GKE
    # validates accepted modes per version). Immutable.
    dataplane_optimization_mode = optional(string, "")

    # Issues a legacy client certificate for control-plane authentication.
    # Off on modern clusters — certificate auth bypasses IAM and cannot be
    # revoked short of rotating the cluster CA; leave unset unless a legacy
    # client genuinely requires it.
    issue_client_certificate = optional(bool)

    # How nodes register themselves: VIA_KUBELET (nodes self-register, the
    # classic path) or VIA_CONTROL_PLANE (the control plane creates node
    # objects — hardens against node impersonation). Immutable.
    node_creation_mode = optional(string, "")

    # Auto-upgrade patch cadence: ACCELERATED upgrades to the latest patch
    # available in the cluster's minor and channel as soon as it ships
    # (instead of the channel's default rollout pacing).
    gke_auto_upgrade_patch_mode = optional(string, "")

    # Locks down the legacy RBAC bindings to system:authenticated /
    # system:unauthenticated — the hardening that prevents accidentally
    # granting cluster access to every Google account on earth.
    rbac_binding_config = optional(object({
      # Allow ClusterRoleBindings/RoleBindings that grant to
      # system:authenticated (every Google-authenticated identity). Leave
      # false — granting to system:authenticated is almost always a mistake.
      enable_insecure_binding_system_authenticated = optional(bool)

      # Allow bindings that grant to system:unauthenticated /
      # system:anonymous. Leave false.
      enable_insecure_binding_system_unauthenticated = optional(bool)
    }))

    # Autopilot-only conversion/posture policies (standard-pool
    # prohibition, system mutation/impersonation guards, webhook safety).
    autopilot_policy = optional(object({
      # Prohibit standard node pools — the cluster runs pure Autopilot.
      no_standard_node_pools = optional(bool)

      # Disallow impersonating system identities.
      no_system_impersonation = optional(bool)

      # Disallow mutating system-managed objects.
      no_system_mutation = optional(bool)

      # Disallow admission webhooks that intercept system-critical requests.
      no_unsafe_webhooks = optional(bool)
    }))

    # Autopilot-only: Cloud Storage / GKE allowlist paths authorizing
    # PRIVILEGED workloads on Autopilot (partner agents etc.). Entries must
    # start with gke:// or gs://; [] allows the default partner allowlists.
    autopilot_privileged_admission_paths = optional(list(string), [])

    # Node settings GKE applies to the pools IT manages on an Autopilot
    # cluster (network tags, Resource Manager tags, kubelet/OS hardening) —
    # the Autopilot counterpart of per-pool node_config.
    node_pool_auto_config = optional(object({
      # GCE network tags applied to Autopilot-managed nodes — what VPC
      # firewall rules match.
      network_tags = optional(list(string), [])

      # Resource Manager tags bound to Autopilot-managed node VMs, as
      # {"tagKeys/123": "tagValues/456"} pairs.
      resource_manager_tags = optional(map(string), {})

      # Container runtime cgroup mode on Autopilot-managed nodes.
      cgroup_mode = optional(string, "")

      # Signed-kernel-module enforcement on Autopilot-managed nodes.
      node_kernel_module_loading_policy = optional(string, "")

      # The kubelet's insecure read-only port 10255 on Autopilot-managed
      # nodes: FALSE closes it (hardened), TRUE keeps it open for legacy
      # agents.
      insecure_kubelet_readonly_port_enabled = optional(string, "")
    }))

    # Defaults inherited by every node pool at CREATION time on a Standard
    # cluster (image streaming, kubelet read-only port, logging variant,
    # containerd registry access). A pool's own node_config overrides
    # these.
    node_pool_defaults = optional(object({
      # Image streaming (GCFS) default for new pools: containers start before
      # the full image is pulled.
      gcfs_enabled = optional(bool)

      # Default posture of the kubelet's insecure read-only port 10255 on new
      # pools: FALSE closes it (hardened), TRUE keeps it open.
      insecure_kubelet_readonly_port_enabled = optional(string, "")

      # Default node system-log throughput for new pools: DEFAULT (100 KiB/s)
      # or MAX_THROUGHPUT (10 MiB/s).
      logging_variant = optional(string, "")

      # Default containerd registry configuration for new pools: private-CA
      # registry trust, per-registry host overrides, writable cgroups.
      containerd_config = optional(object({
        # Trust custom certificate authorities for specific registry domains.
        private_registry_access = optional(object({
          # Master toggle for private registry access configuration.
          enabled = optional(bool, false)

          # Per-domain CA trust entries.
          certificate_authority_domains = optional(list(object({
            # Registry FQDNs this CA vouches for, e.g. ["registry.internal:5000"].
            fqdns = list(string)

            # Secret Manager secret URI holding the CA certificate, in the form
            # "projects/{project}/secrets/{secret}/versions/{version}".
            gcp_secret_manager_certificate_uri = string
          })), [])
        }))

        # Per-registry host overrides (mirrors, capabilities, auth, headers).
        registry_hosts = optional(list(object({
          # The registry server the overrides apply to, e.g. "docker.io".
          server = string

          # Host endpoints serving this registry (mirrors first, in order).
          hosts = optional(list(object({
            # Endpoint URL, e.g. "https://mirror.internal".
            host = string

            # Operations this endpoint can serve — typically "pull" and "resolve"
            # (containerd hosts.toml capability names).
            capabilities = optional(list(string), [])

            # Dial timeout for this endpoint, e.g. "10s".
            dial_timeout = optional(string, "")

            # Path override on the endpoint host.
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

            # Custom HTTP headers sent to this endpoint.
            headers = optional(map(string), {})
          })), [])
        })), [])

        # Writable cgroup filesystem inside containers.
        writable_cgroups_enabled = optional(bool)
      }))
    }))

    # Bring-your-own control-plane keys: customer-managed CAs for the
    # cluster/etcd/aggregation trust domains and KMS keys for control-plane
    # disk encryption and ServiceAccount JWT signing. For regulated
    # environments that must own the entire trust chain. Immutable.
    user_managed_keys = optional(object({
      # CA Service CaPool issuing the cluster CA
      # ("projects/{p}/locations/{l}/caPools/{pool}").
      cluster_ca = optional(string, "")

      # CA Service CaPool for the etcd API CA.
      etcd_api_ca = optional(string, "")

      # CA Service CaPool for the etcd peer CA.
      etcd_peer_ca = optional(string, "")

      # CA Service CaPool for the aggregation layer CA.
      aggregation_ca = optional(string, "")

      # KMS key encrypting the control-plane disks
      # ("projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{k}"). Accepts a
      # literal path or a reference to a GcpKmsKey resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      control_plane_disk_encryption_key = optional(string, "")

      # KMS key encrypting GKE-ops etcd backups. Accepts a literal path or a
      # reference to a GcpKmsKey resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      gkeops_etcd_backup_encryption_key = optional(string, "")

      # KMS cryptoKeyVersions SIGNING ServiceAccount JWTs issued by the
      # cluster ("projects/.../cryptoKeyVersions/1").
      service_account_signing_keys = optional(list(string), [])

      # KMS cryptoKeyVersions accepted for VERIFYING ServiceAccount JWTs
      # (include the previous version during signing-key rotation).
      service_account_verification_keys = optional(list(string), [])
    }))

    # Rotation of Secret Manager secrets mounted through the built-in CSI
    # add-on (requires enable_secret_manager_csi): re-fetch cadence for
    # mounted secret values.
    secret_manager_rotation = optional(object({
      # Whether mounted secrets are periodically re-fetched.
      enabled = optional(bool, false)

      # Re-fetch cadence, seconds format (e.g. "120s"; default 120s).
      rotation_interval = optional(string, "")
    }))

    # The Secret Manager SYNC add-on: syncs Secret Manager secrets into
    # Kubernetes Secret objects (as opposed to the CSI add-on's volume
    # mounts), with its own rotation cadence.
    secret_sync = optional(object({
      # Whether the sync add-on is enabled.
      enabled = optional(bool, false)

      # Whether synced secrets are periodically refreshed.
      rotation_enabled = optional(bool, false)

      # Refresh cadence, seconds format (e.g. "120s").
      rotation_interval = optional(string, "")
    }))

    # Two-step (rollback-safe) control-plane minor upgrades: the control
    # plane moves to the new version but keeps emulating the old minor for
    # a soak period during which the upgrade can be rolled back without
    # data loss. Leave unset for standard one-step upgrades.
    rollback_safe_upgrade = optional(object({
      # How long the cluster stays in the rollbackable state after the
      # control plane upgrades, as a seconds-format duration, e.g. "604800s"
      # (7 days). Minimum 6 hours ("21600s"), maximum 7 days ("604800s").
      # Leave empty to skip the two-step flow and perform a standard
      # one-step upgrade.
      # The bound is expressed as one pattern (21600 <= seconds <= 604800)
      # so every validation engine, including the Java one, evaluates it
      # without string slicing.
      control_plane_soak_duration = optional(string, "")
    }))

    # Completes a rollback-safe upgrade declaratively: set to the target
    # minor version ("major.minor", e.g. "1.33") once the soak period has
    # proven the new control plane, and GKE stops emulating the old minor.
    # Removing the field does not trigger completion; only setting it does.
    # Only meaningful with rollback_safe_upgrade.
    desired_emulated_version = optional(string, "")
  })
}
