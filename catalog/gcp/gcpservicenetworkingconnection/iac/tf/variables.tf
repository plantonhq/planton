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
  description = "GcpServiceNetworkingConnection specification"
  type = object({
    # The GCP project used to enable the required APIs
    # (servicenetworking.googleapis.com and compute.googleapis.com) before the
    # connection is created. Can be a literal project ID or a reference to a
    # GcpProject resource. If omitted, the provider's default project is used.
    # The connection itself is addressed by the network (GCP derives the
    # owning project from it) — set this only when the network's project
    # differs from the provider's default project.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The VPC network to peer with the service producer. Accepts a network
    # name or full self-link URL; a reference resolves to the GcpVpcNetwork's
    # self-link. Immutable: changing it destroys and recreates the connection
    # (and with it, every producer resource's private connectivity).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    network = string

    # The service producer to peer with (default
    # servicenetworking.googleapis.com — the producer behind Cloud SQL,
    # AlloyDB, Memorystore, and Filestore private IP). Third-party producers
    # publish their own service names. Immutable: one connection exists per
    # (network, service) pair.
    service = optional(string, "")

    # Named INTERNAL VPC_PEERING address ranges (GcpGlobalAddress resources,
    # referenced by NAME — not self-link or CIDR) reserved for the producer to
    # allocate service subnets from. At least one range is required. Mutable:
    # append ranges here when the producer runs out of space — GCP keeps
    # already-provisioned producer subnets even when the list changes, so
    # growth is additive and safe. Size ranges generously up front (a /16 is
    # the common default): producers cannot use ranges that are too fragmented.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    reserved_peering_ranges = list(string)

    # Recovery lever for a known API wrinkle: when a connection for this
    # (network, service) pair already exists outside of Planton's management
    # (e.g. created by gcloud or a console flow), the create call fails with
    # "Cannot modify allocated ranges". Setting this to true converts that
    # failure into an in-place update of the existing connection's reserved
    # ranges, adopting it instead of erroring. Leave false unless you are
    # deliberately taking over a pre-existing connection.
    update_on_creation_fail = optional(bool, false)

    # Deletion policy for the connection — what happens when this resource
    # is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the peering is deleted (GCP refuses while any service
    #                producer still holds subnets in the reserved ranges —
    #                destroy the Cloud SQL / AlloyDB / Memorystore instances
    #                first; see the teardown note above)
    #   "PREVENT" -- destroy FAILS; protects the private connectivity every
    #                producer instance on this network depends on
    #   "ABANDON" -- the connection is removed from management but keeps
    #                serving private IPs in GCP — the historical workaround
    #                for stuck teardowns, at the cost of an unmanaged peering
    deletion_policy = optional(string, "")
  })
}
