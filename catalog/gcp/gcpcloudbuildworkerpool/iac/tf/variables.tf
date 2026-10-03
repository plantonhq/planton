variable "metadata" {
  description = "Catalog object metadata"
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
  description = "GcpCloudBuildWorkerPool specification"
  type = object({
    # The project the pool lives in: a literal project ID or a GcpProject
    # reference. Empty means the provider's default project. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The region the pool's workers run in, e.g. "us-central1". Builds that
    # use the pool run here, and a trigger naming it must be in the same
    # region. Required. Immutable.
    location = string

    # The pool's ID, unique in the project and region. Defaults to
    # metadata.name. Immutable.
    worker_pool_id = optional(string, "")

    # A human-readable name shown in the console, 1-63 characters.
    display_name = optional(string, "")

    # Annotations on the pool (Google's AIP-128 key/value metadata; not
    # labels -- the pool has none). Only the keys declared here are managed.
    annotations = optional(map(string), {})

    # Peer the pool's workers into one of your VPC networks. Requires a
    # Service Networking (private services access) connection on that
    # network. Mutually exclusive with private_service_connect. Immutable.
    network_config = optional(object({
      # The VPC network the workers are peered to: a GcpVpcNetwork reference
      # or a literal projects/{project}/global/networks/{name}. Google requires
      # the project NUMBER in that path; both modules resolve a project ID to
      # its number (one project lookup at plan time). The network must already
      # have private services access configured. Required. Immutable.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      peered_network = string

      # The subnet range within the peered network the workers take addresses
      # from, in CIDR notation: a prefix alone ("/26") lets Google pick the
      # block, an address with a prefix ("192.168.0.0/29") pins it. Empty
      # means "/24". Immutable.
      peered_network_ip_range = optional(string, "")
    }))

    # Attach each worker to a Private Service Connect network attachment in
    # the pool's region. Mutually exclusive with network_config. Immutable.
    private_service_connect = optional(object({
      # The network attachment each worker's interface connects to, as
      # projects/{project}/regions/{region}/networkAttachments/{name}, in the
      # pool's region. Required. Immutable.
      network_attachment = string

      # Route ALL worker traffic through the PSC interface, for full control
      # of egress (configure Cloud NAT on the attachment's subnet to reach the
      # internet). False routes only private ranges (10.0.0.0/8,
      # 172.16.0.0/12, 192.168.0.0/16) through it. Immutable.
      route_all_traffic = optional(bool, false)
    }))

    # The workers' machine shape and public addressing. Omitted fields keep
    # Cloud Build's defaults (n1-standard-1, a standard disk, public IPs).
    worker_config = optional(object({
      # The worker's machine type, e.g. "e2-standard-4" or "n1-highcpu-8".
      # Empty means n1-standard-1.
      machine_type = optional(string, "")

      # The worker's disk size in GB, up to 1000. 0 means Cloud Build's
      # standard disk size.
      disk_size_gb = optional(number, 0)

      # Run workers without public IP addresses, which blocks egress to public
      # IPs (pair with network_config or private_service_connect and Cloud NAT
      # when builds still need the internet). Unset keeps Cloud Build's
      # default (public IPs); set false explicitly to restore them.
      no_external_ip = optional(bool)

      # Enable nested virtualization on the workers, for builds that run VMs
      # or emulators, when the machine type supports it. Unset keeps Cloud
      # Build's default (off).
      enable_nested_virtualization = optional(bool)
    }))

    # What destroy does:
    #   "" / "DELETE" -- the pool is deleted
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the pool leaves management and stays in Google
    deletion_policy = optional(string, "")
  })
}
