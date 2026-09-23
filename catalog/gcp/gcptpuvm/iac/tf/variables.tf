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
  description = "GcpTpuVm specification"
  type = object({
    # The GCP project the TPU lives in: a literal project ID or a GcpProject
    # reference. If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The zone the TPU runs in, e.g. "us-central1-a" or "europe-west4-a".
    # Only some zones offer each TPU generation; check Google's TPU regions
    # and zones list. Immutable.
    zone = string

    # The TPU's id -- its name in the project. Lowercase letters, digits,
    # and hyphens, starting with a letter. Defaults to metadata.name.
    # Immutable.
    node_id = optional(string, "")

    # The TPU software image, matched to the accelerator generation and
    # framework, e.g. "tpu-ubuntu2204-base", "v2-alpha-tpuv5-lite",
    # "v2-alpha-tpuv6e". Immutable.
    runtime_version = string

    # The slice by name: generation and chip or core count, e.g. "v2-8",
    # "v3-8", "v5litepod-8", "v6e-8". Set this or accelerator_config; with
    # neither, Google's provider asks for "v2-8". Immutable.
    accelerator_type = optional(string, "")

    # The slice by generation and topology -- how you ask for shapes the
    # type names do not cover. Set this or accelerator_type. Immutable.
    accelerator_config = optional(object({
      # The TPU generation: "V2", "V3", "V4", "V5LITE_POD", "V5P", or "V6E".
      type = string

      # The chip topology, e.g. "2x2" or "2x2x1".
      topology = string
    }))

    # What the TPU is for.
    description = optional(string, "")

    # A /29 CIDR block the TPU picks its address from. It must not overlap
    # any subnetwork of the network, any peered network, or another TPU's
    # block. Unset: Google chooses. Immutable.
    cidr_block = optional(string, "")

    # The TPU's network. Set this or network_configs. Unset: the project's
    # default network and subnetwork. Immutable.
    network_config = optional(object({
      # The VPC network: a GcpVpcNetwork reference or a literal
      # projects/{project}/global/networks/{name}. Unset: "default".
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      network = optional(string, "")

      # The subnetwork in the TPU's region: a GcpSubnetwork reference or a
      # literal projects/{project}/regions/{region}/subnetworks/{name}. Unset:
      # "default".
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnetwork = optional(string, "")

      # Give the TPU workers external IP addresses. Without them the
      # subnetwork needs Private Google Access (or Cloud NAT) to reach
      # Google APIs and the internet.
      enable_external_ips = optional(bool, false)

      # Let the workers send and receive packets with non-matching source or
      # destination addresses -- needed only when they forward routes.
      can_ip_forward = optional(bool, false)

      # The number of queues on the interface (higher network throughput on
      # large hosts). Unset: Google's default.
      queue_count = optional(number, 0)
    }))

    # Several network interfaces, one per entry, for multi-NIC TPU VMs.
    # Set this or network_config. Immutable.
    network_configs = optional(list(object({
      # The VPC network: a GcpVpcNetwork reference or a literal
      # projects/{project}/global/networks/{name}. Unset: "default".
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      network = optional(string, "")

      # The subnetwork in the TPU's region: a GcpSubnetwork reference or a
      # literal projects/{project}/regions/{region}/subnetworks/{name}. Unset:
      # "default".
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnetwork = optional(string, "")

      # Give the TPU workers external IP addresses. Without them the
      # subnetwork needs Private Google Access (or Cloud NAT) to reach
      # Google APIs and the internet.
      enable_external_ips = optional(bool, false)

      # Let the workers send and receive packets with non-matching source or
      # destination addresses -- needed only when they forward routes.
      can_ip_forward = optional(bool, false)

      # The number of queues on the interface (higher network throughput on
      # large hosts). Unset: Google's default.
      queue_count = optional(number, 0)
    })), [])

    # The identity the TPU host VMs run as. Unset: the Compute Engine
    # default service account with access to every Cloud API. Immutable.
    service_account = optional(object({
      # The service account: a GcpServiceAccount reference or a literal email.
      # Unset: the Compute Engine default service account.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      email = optional(string, "")

      # OAuth scopes granted to it. Unset: every Cloud API
      # (https://www.googleapis.com/auth/cloud-platform); access is then
      # governed by the account's IAM roles.
      scopes = optional(list(string), [])
    }))

    # Spot, preemptible, or reserved capacity. Unset: on-demand. Immutable.
    scheduling_config = optional(object({
      # Preemptible capacity (the older model; Google may reclaim it and ends
      # it after 24 hours).
      preemptible = optional(bool, false)

      # Spot capacity -- the cheapest; Google may reclaim it at any time, with
      # no 24-hour limit. Checkpoint often.
      spot = optional(bool, false)

      # Draw from a reservation made for this project and zone.
      reserved = optional(bool, false)
    }))

    # Existing persistent disks to attach -- training data or checkpoints
    # shared across TPUs. Mutable.
    data_disks = optional(list(object({
      # The disk: a GcpComputeDisk reference or a literal
      # projects/{project}/zones/{zone}/disks/{name} in the TPU's zone.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      source_disk = string

      # "READ_WRITE" (the default; one TPU at a time) or "READ_ONLY" (shared
      # read access for many TPUs).
      mode = optional(string, "")
    })), [])

    # Boot the host VMs with Secure Boot (Shielded VM). Sent only when true.
    # Immutable.
    enable_secure_boot = optional(bool, false)

    # Labels on the TPU. The platform attribution labels are merged in and
    # win on key conflicts.
    labels = optional(map(string), {})

    # Custom metadata on the host VMs -- for example "startup-script" (runs
    # on every worker at boot) and "shutdown-script". Mutable.
    metadata = optional(map(string), {})

    # Network tags on the host VMs -- the handle VPC firewall rules target.
    # Mutable.
    tags = optional(list(string), [])

    # What happens to the TPU when this resource is destroyed:
    #   "" / "DELETE" -- the TPU and its host VMs are deleted (attached data
    #                    disks are only detached)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the TPU leaves management and keeps running (and
    #                    billing)
    deletion_policy = optional(string, "")
  })
}
