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
  description = "GcpDatastreamPrivateConnection specification"
  type = object({
    # The GCP project the private connection lives in: a literal project ID
    # or a GcpProject reference. If omitted, the provider's default project
    # is used. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Datastream region, e.g. "us-central1". Profiles and streams that
    # use this connection live in the same region. Immutable.
    location = string

    # The private connection's ID. Defaults to metadata.name. Immutable.
    private_connection_id = optional(string, "")

    # The name shown in the console. Defaults to metadata.name. Immutable.
    display_name = optional(string, "")

    # Peer Datastream's network with a VPC.
    vpc_peering_config = optional(object({
      # The VPC Datastream peers with, as projects/{project}/global/networks/{name}
      # -- a GcpVpcNetwork reference (its network_id output) or a literal in
      # that form. Immutable.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      vpc = string

      # A free /29 range in the VPC's address space for Datastream's side of
      # the peering, e.g. "10.200.0.0/29" -- it must overlap no subnet, no
      # other peering, and no range routed into the VPC. Immutable.
      subnet = string
    }))

    # Reach a VPC through a Private Service Connect interface.
    psc_interface_config = optional(object({
      # The network attachment, as
      # projects/{project}/regions/{region}/networkAttachments/{name}, in the
      # private connection's region. It must accept connections from
      # Datastream's service project. Immutable.
      network_attachment = string
    }))

    # Skip Google's checks at create (for example, that the /29 is free).
    # Useful when the network is still being built; a real conflict then
    # surfaces when a profile first connects. Immutable.
    create_without_validation = optional(bool, false)

    # Labels on the private connection. The platform attribution labels are
    # added on top and win on a key conflict.
    labels = optional(map(string), {})

    # What happens to the private connection when this resource is
    # destroyed:
    #   "" / "FORCE" -- deleted together with the routes Datastream created
    #                   on it (the provider's default)
    #   "DELETE"     -- deleted without force; fails while routes remain
    #   "PREVENT"    -- destroy fails
    #   "ABANDON"    -- it leaves management and stays in GCP
    # Delete the profiles that use it first; Google refuses to delete a
    # connection a profile still references.
    deletion_policy = optional(string, "")
  })
}
