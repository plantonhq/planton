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
  description = "GcpVertexAiPersistentResource specification"
  type = object({
    # The GCP project the resource lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Vertex AI location (region) the resource runs in, e.g.
    # "us-central1". Jobs that use it must run in the same location.
    # Immutable.
    location = string

    # The resource's id -- what a job's persistent_resource_id names. Up to
    # 63 characters: lowercase letters, digits, and hyphens, starting with a
    # letter and not ending with a hyphen. Defaults to metadata.name.
    # Immutable.
    persistent_resource_id = optional(string, "")

    # Human-readable name shown in the console -- up to 128 UTF-8
    # characters. Mutable in place.
    display_name = optional(string, "")

    # Labels on the resource. The platform attribution labels are merged in
    # and win on key conflicts. Mutable in place.
    labels = optional(map(string), {})

    # The pools of machines the resource keeps provisioned. At least one.
    resource_pools = list(object({
      # The pool's id within the resource -- what a job's worker pool spec
      # refers to. Google generates one when empty. Sent only when set.
      # Immutable.
      id = optional(string, "")

      # The machine every replica runs on.
      machine_spec = object({
        # The Compute Engine machine type, e.g. "n1-standard-4", "a2-highgpu-1g",
        # "ct5lp-hightpu-4t" -- any type Vertex AI custom training supports.
        # Immutable.
        machine_type = optional(string, "")

        # The accelerator attached to every replica (NVIDIA_TESLA_T4,
        # NVIDIA_L4, NVIDIA_TESLA_A100, NVIDIA_A100_80GB, NVIDIA_H100_80GB,
        # NVIDIA_H100_MEGA_80GB, NVIDIA_H200_141GB, NVIDIA_B200, NVIDIA_GB200,
        # NVIDIA_RTX_PRO_6000, the older NVIDIA_TESLA_K80 / P100 / V100 / P4,
        # or TPU_V2 / TPU_V3 / TPU_V4_POD / TPU_V5_LITEPOD). The machine type
        # must be one Google pairs with it. Empty attaches none. Immutable.
        accelerator_type = optional(string, "")

        # How many accelerators each replica carries, set together with
        # accelerator_type. Sent only when set. Immutable.
        accelerator_count = optional(number, 0)
      })

      # How many replicas the pool runs (and bills) whether or not a job is
      # using them. Sent as a decimal string. Mutable in place.
      replica_count = optional(number)

      # Autoscaling bounds for the pool. Omit for a fixed replica_count.
      autoscaling_spec = optional(object({
        # Fewest replicas kept running -- at least 1 on a persistent resource
        # (Google rejects 0 here), and at most the pool's replica_count. Sent as
        # a decimal string. Immutable.
        min_replica_count = optional(number)

        # Most replicas the pool may scale to -- more than min_replica_count and
        # at least the pool's replica_count. Sent as a decimal string.
        # Immutable.
        max_replica_count = optional(number)
      }))

      # Boot disk options. Omit for Google's defaults. Sent only when set.
      disk_spec = optional(object({
        # Boot disk size in GB. Google defaults to 100. Sent only when set.
        # Immutable.
        boot_disk_size_gb = optional(number)

        # Boot disk type: "pd-ssd" (Google's default on most machines),
        # "pd-standard", or "hyperdisk-balanced" (the default on A3 Ultra). Sent
        # only when set. Immutable.
        boot_disk_type = optional(string, "")
      }))
    }))

    # A VPC network to peer the resource with, so jobs reach private
    # services: a GcpVpcNetwork reference or a literal
    # projects/{project}/global/networks/{name}. Google requires the project
    # NUMBER in that path; both modules resolve a project ID to its number
    # (one project lookup at plan time). The network must already have VPC
    # Network Peering for Vertex AI (private services access) configured.
    # Omit for no peering. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    network = optional(string, "")

    # Names of the private-services-access ranges on the peered network the
    # resource's machines take addresses from, e.g. ["vertex-ai-range"]
    # (the name of a GcpGlobalAddress with purpose VPC_PEERING). Empty lets
    # Google use any range allocated to the peering. Requires network.
    # Immutable.
    reserved_ip_ranges = optional(list(string), [])

    # A Private Service Connect interface into your VPC -- the alternative
    # to network (peering); Google's documentation treats the two as
    # alternatives and its API is the authority when both are set.
    # Immutable.
    psc_interface_config = optional(object({
      # The Compute Engine network attachment in the resource's region the
      # interface joins, as its name or its full
      # projects/{project}/regions/{region}/networkAttachments/{name} path.
      # Create the attachment first. Immutable.
      network_attachment = optional(string, "")

      # Domains Google's tenant VPC resolves through Cloud DNS zones in your
      # networks.
      dns_peering_configs = optional(list(object({
        # The DNS suffix peered to, ending with a dot, e.g.
        # "corp.example.com.". Immutable.
        domain = string

        # The project hosting the Cloud DNS zone for the domain: a GcpProject
        # reference or a literal project ID. The Vertex AI Service Agent needs
        # roles/dns.peer on it. Immutable.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        target_project = string

        # The VPC network, in target_project, where the zone is visible: a
        # GcpVpcNetwork reference (its name) or a literal network name.
        # Immutable.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        target_network = string
      })), [])
    }))

    # True requires every job on the resource to run as a custom,
    # user-managed service account (the job names it); false runs jobs as
    # the Vertex AI Custom Code Service Agent. Immutable.
    enable_custom_service_account = optional(bool, false)

    # Customer-managed encryption key protecting the resource's disks: a
    # GcpKmsKey reference or a literal
    # projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
    # in the same region. Jobs on the resource must use the same key. Omit to
    # use Google-managed encryption. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # What happens to the resource when this block is destroyed:
    #   "" / "DELETE" -- the machines are released and billing stops
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the resource leaves management and keeps running
    #                    (and billing)
    deletion_policy = optional(string, "")
  })
}
