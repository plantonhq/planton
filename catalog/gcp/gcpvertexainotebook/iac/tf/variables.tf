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
  description = "GcpVertexAiNotebook specification"
  type = object({
    # GCP project where the notebook instance will be created.
    # If omitted, the instance is created in the provider's default
    # project (from the credential or ambient configuration).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # GCP zone where the notebook instance will be created (e.g., "us-central1-a").
    # Immutable after creation.
    location = string

    # Compute Engine machine type for the instance.
    # Examples: "e2-standard-4", "n1-standard-8", "a2-highgpu-1g".
    # Choose based on workload requirements -- CPU-only for data processing,
    # N1/A2 for GPU-accelerated ML training.
    machine_type = string

    # Name of the Workbench instance in GCP.
    # If not specified, defaults to metadata.name.
    # Immutable after creation. Must be a valid RFC1035 hostname.
    instance_name = optional(string, "")

    # Email addresses of users who own the instance.
    # Format: alias@example.com. Currently GCP supports one owner only.
    # If set, access mode is Single User (only this owner can access).
    # Immutable after creation.
    instance_owners = optional(list(string), [])

    # Desired state of the instance: ACTIVE (running) or STOPPED.
    # Use STOPPED to suspend the instance and stop billing for compute
    # (storage charges still apply). Defaults to ACTIVE.
    desired_state = optional(string, "")

    # If true, the notebook instance will not register with the proxy
    # and no JupyterLab proxy URL will be generated. Use this for
    # instances accessed only via SSH or other direct methods.
    # Immutable after creation.
    disable_proxy_access = optional(bool, false)

    # Custom metadata key-value pairs for the instance.
    # Some keys trigger special behaviors (e.g., install-monitoring-agent).
    metadata = optional(map(string), {})

    # User-defined labels to organize the instance (cost attribution,
    # team ownership, environment tagging). Keys and values must follow
    # GCP label rules: lowercase letters, digits, underscores, and dashes,
    # at most 63 characters. Merged with the platform's attribution labels;
    # on key conflicts the platform labels win. Mutable in place.
    labels = optional(map(string), {})

    # Boot disk configuration for the notebook instance.
    # If not specified, GCP provisions a 150 GB PD_SSD boot disk with
    # Google-managed encryption.
    boot_disk = optional(object({
      # Disk type for the boot disk.
      # Persistent Disk: PD_STANDARD (HDD), PD_SSD, PD_BALANCED, PD_EXTREME.
      # Hyperdisk (newer-generation machine series only): HYPERDISK_BALANCED,
      # HYPERDISK_BALANCED_HIGH_AVAILABILITY, HYPERDISK_ML.
      # If not specified, defaults to PD_SSD.
      disk_type = optional(string, "")

      # Size of the boot disk in GB.
      # Minimum 10 GB, maximum 64000 GB (64 TB).
      # If not specified, defaults to 150 GB.
      disk_size_gb = optional(number, 0)

      # KMS key for CMEK encryption of the boot disk.
      # Format: projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
      # If not specified, Google-managed encryption (GMEK) is used.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key = optional(string, "")
    }))

    # Data disk configuration for the notebook instance.
    # GCP supports exactly one data disk per Workbench instance.
    # If not specified, GCP provisions a 100 GB data disk with
    # Google-managed encryption.
    data_disk = optional(object({
      # Disk type for the data disk.
      # Persistent Disk: PD_STANDARD (HDD), PD_SSD, PD_BALANCED, PD_EXTREME.
      # Hyperdisk (newer-generation machine series only): HYPERDISK_BALANCED,
      # HYPERDISK_EXTREME, HYPERDISK_THROUGHPUT,
      # HYPERDISK_BALANCED_HIGH_AVAILABILITY, HYPERDISK_ML.
      # If not specified, defaults to PD_STANDARD.
      disk_type = optional(string, "")

      # Size of the data disk in GB.
      # Minimum 10 GB, maximum 64000 GB (64 TB).
      # If not specified, defaults to 100 GB.
      disk_size_gb = optional(number, 0)

      # KMS key for CMEK encryption of the data disk.
      # Format: projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
      # If not specified, Google-managed encryption (GMEK) is used.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key = optional(string, "")

      # Compute Engine resource policies attached to the data disk — most
      # usefully a snapshot schedule, so the notebook's working data is
      # backed up on a cadence without any agent inside the VM. Each entry is
      # a policy's full resource name or self link
      # (projects/{project}/regions/{region}/resourcePolicies/{name}); the
      # policy must live in the instance's region. Leave empty for no
      # attached policies; sent only when set because the API reports the
      # attached set itself.
      resource_policies = optional(list(string), [])
    }))

    # GPU accelerator configuration for the notebook instance.
    # GCP supports one accelerator configuration per instance.
    # Requires compatible machine types (e.g., n1-standard-* for Tesla GPUs).
    accelerator_config = optional(object({
      # GPU accelerator type. Availability varies by zone and machine type:
      # current-generation training/inference GPUs (NVIDIA_H100_80GB,
      # NVIDIA_H100_MEGA_80GB, NVIDIA_H200_141GB, NVIDIA_B200) require
      # their matching A3/A4 machine series; NVIDIA_L4 and the Tesla
      # family run on G2/N1; the _VWS variants are virtual-workstation
      # (graphics) licenses of the same silicon.
      # See https://cloud.google.com/vertex-ai/docs/workbench/instances/create#accelerator
      # for supported types per zone.
      type = optional(string, "")

      # Number of accelerator cores.
      # Valid values depend on the accelerator type (typically 1, 2, 4, or 8).
      core_count = optional(number, 0)
    }))

    # Network interface configuration for the notebook instance.
    # GCP supports one network interface per instance.
    # If not specified, the instance uses the default VPC network.
    # Immutable after creation.
    network_interface = optional(object({
      # VPC network for the instance.
      # Can be a literal value (VPC name or self_link) or a reference to
      # a GcpVpcNetwork resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      network = optional(string, "")

      # Subnetwork for the instance.
      # Can be a literal value (subnet name or self_link) or a reference to
      # a GcpSubnetwork resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnet = optional(string, "")

      # NIC type for the network interface.
      # Valid values: VIRTIO_NET, GVNIC.
      # GVNIC provides higher bandwidth and lower latency.
      # If not specified, defaults to VIRTIO_NET.
      nic_type = optional(string, "")

      # Static external IP address for the instance (ONE_TO_ONE_NAT access
      # config). The address must be an unused static external IP in the
      # same region as the instance's zone. Reference a GcpAddress resource
      # to reserve and pin the IP as a first-class node, or pass a literal
      # IP address.
      #
      # If omitted (and public IP is not disabled), GCP assigns an ephemeral
      # external IP from a shared pool -- the IP changes across stop/start
      # cycles. Pin a static address when firewall allowlists or DNS records
      # depend on the instance's IP.
      #
      # Cannot be combined with disable_public_ip. Immutable after creation.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      external_ip = optional(string, "")
    }))

    # If true, no external IP is assigned to the instance.
    # Use this for instances that should only be accessible through
    # the Vertex AI proxy or a VPN/Cloud IAP tunnel.
    # Immutable after creation.
    disable_public_ip = optional(bool, false)

    # If true, enable IP forwarding on the instance. Useful for
    # instances acting as network gateways. Default false.
    # Immutable after creation.
    enable_ip_forwarding = optional(bool, false)

    # Service account email for the notebook VM identity.
    # The VM uses this service account to access GCP resources
    # (BigQuery, GCS, Vertex AI, etc.). Scopes are fixed to
    # "https://www.googleapis.com/auth/cloud-platform".
    # Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    service_account = optional(string, "")

    # Compute Engine network tags for firewall rules.
    # Immutable after creation.
    tags = optional(list(string), [])

    # VM image configuration for the notebook environment.
    # Uses pre-built deep learning VM images from GCP with popular
    # ML frameworks (TensorFlow, PyTorch, JAX) and JupyterLab.
    # Mutually exclusive with container_image.
    # Immutable after creation.
    vm_image = optional(object({
      # Google Cloud project that the VM image belongs to.
      # GCP's own Workbench images live in "cloud-notebooks-managed".
      # (The legacy deep-learning-VM notebook families under
      # "deeplearning-platform-release" have been retired by GCP and no
      # longer resolve.)
      project = optional(string, "")

      # VM image family. The newest image in this family will be used.
      # GCP's maintained family for Workbench instances is
      # "workbench-instances" (in the "cloud-notebooks-managed" project).
      # Mutually exclusive with name (within this message).
      family = optional(string, "")

      # Specific VM image name. Use this instead of family when you need
      # to pin to an exact image version.
      # Mutually exclusive with family (within this message).
      name = optional(string, "")
    }))

    # Container image configuration for a custom notebook environment.
    # Use this when pre-built VM images don't meet your needs.
    # Mutually exclusive with vm_image.
    container_image = optional(object({
      # Container image repository path.
      # Example: "gcr.io/deeplearning-platform-release/base-cu113.py310"
      repository = string

      # Container image tag. If not specified, the latest tag is used.
      tag = optional(string, "")
    }))

    # Shielded VM configuration for enhanced security.
    # Shielded VMs protect against rootkits and bootkits with
    # Secure Boot, vTPM, and integrity monitoring.
    shielded_instance_config = optional(object({
      # Enable Secure Boot. Ensures only verified boot software runs.
      # Disabled by default because some ML libraries may not have signed
      # boot loaders.
      enable_secure_boot = optional(bool)

      # Enable vTPM (Virtual Trusted Platform Module).
      # Provides measured boot integrity and key generation.
      # Enabled by default; an explicit false turns it off.
      enable_vtpm = optional(bool)

      # Enable integrity monitoring. Compares boot measurements against
      # a trusted baseline.
      # Enabled by default; an explicit false turns it off.
      enable_integrity_monitoring = optional(bool)
    }))

    # Confidential Computing configuration. When set, guest memory is
    # encrypted in use with AMD SEV -- data stays protected even from the
    # host hypervisor. Requires an SEV-capable AMD machine type (e.g., the
    # n2d family). Immutable after creation.
    confidential_instance_config = optional(object({
      # Confidential computing technology for the instance.
      # The only supported value is SEV (AMD Secure Encrypted Virtualization).
      # If not specified, defaults to SEV.
      confidential_instance_type = optional(string, "")

      # Whether Confidential Computing is on. Unset means on: declaring the
      # block has always meant enabling it, and this switch lets a manifest say
      # the opposite out loud.
      enabled = optional(bool)
    }))

    # Compute Engine reservation affinity. Points the notebook VM at
    # pre-purchased capacity reservations -- the way organizations
    # guarantee GPU availability for ML workloads. Immutable after creation.
    reservation_affinity = optional(object({
      # How the instance consumes reservations:
      #   - RESERVATION_ANY (default): consume any matching open reservation.
      #   - RESERVATION_SPECIFIC: consume only the reservation named in
      #     key/values.
      #   - RESERVATION_NONE: never consume reserved capacity (on-demand only).
      consume_reservation_type = optional(string, "")

      # Corresponds to the label key of a reservation resource. To target a
      # SPECIFIC_RESERVATION by name, use "compute.googleapis.com/reservation-name"
      # as the key. Only valid with RESERVATION_SPECIFIC.
      key = optional(string, "")

      # Corresponds to the label values of a reservation resource -- for the
      # reservation-name key, the reservation's name. Only valid with
      # RESERVATION_SPECIFIC.
      values = optional(list(string), [])
    }))

    # Enable managed end-user credentials (EUC) for the instance. With
    # managed EUC, JupyterLab runs as the accessing user's own Google
    # identity rather than the VM's service account, so notebook code sees
    # the user's IAM permissions (single-user auditability). Mutable in place.
    enable_managed_euc = optional(bool, false)

    # Allow access to the notebook through a third-party identity provider
    # configured on the organization (workforce identity federation).
    # Leave false when all users authenticate with Google identities.
    # Mutable in place.
    enable_third_party_identity = optional(bool, false)

    # Deletion policy for the notebook — what happens when this resource
    # is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the instance is deleted, including its boot and data
    #                disks; unsynced notebook work on those disks is lost
    #   "PREVENT" -- destroy FAILS; a guard for a workstation whose local
    #                disks hold work not yet pushed anywhere else
    #   "ABANDON" -- the instance is removed from management but left
    #                running (and billing) in GCP with its disks intact
    deletion_policy = optional(string, "")

    # Minimum CPU platform for the VM, e.g. "Intel Cascade Lake" or
    # "Intel Sapphire Rapids": pins the instance to at least this CPU
    # generation so notebook code that relies on newer instruction sets
    # (AVX-512, AMX) is never scheduled on older hardware. The platform
    # must be offered for the machine type in the instance's zone. Leave
    # empty to let Compute Engine choose; sent only when set because the
    # API reports the platform it picked.
    min_cpu_platform = optional(string, "")

    # Workbench-side deletion protection: while true, the API refuses to
    # delete the instance from any client (console, gcloud, either IaC
    # engine) until the flag is lifted — a guard for a workstation whose
    # local disks hold work not yet pushed anywhere else. Off unless set;
    # pair with deletion_policy PREVENT for an engine-side guard too.
    # Sent only when set because the API reports its current value.
    enable_deletion_protection = optional(bool)
  })
}
