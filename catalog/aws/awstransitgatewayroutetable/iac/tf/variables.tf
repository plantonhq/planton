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
  description = "AwsTransitGatewayRouteTable specification"
  type = object({
    # The AWS region where the route table will be created. Must match the
    # Transit Gateway's region.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # The Transit Gateway this route table belongs to. Create-time immutable:
    # changing it replaces the table (and everything folded in it). Reference
    # an AwsTransitGateway's transit_gateway_id output or pass a literal ID.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    transit_gateway_id = string

    # Attachments ASSOCIATED with this table: their outbound traffic is looked
    # up here. An attachment can be associated with at most ONE route table
    # across the whole gateway (AWS enforces this at apply time), so an
    # attachment listed here must have its default_route_table_association
    # turned off, must not appear in any other table's associations -- or must
    # be taken over explicitly with the entry's replace_existing_association.
    # Attachment IDs must be unique within the table (both engines key each
    # association by its attachment ID).
    associations = optional(list(object({
      # The attachment whose outbound traffic is looked up in this table.
      # Reference an AwsTransitGatewayVpcAttachment's attachment_id output, or
      # pass a literal attachment ID for VPN / Direct Connect / peering
      # attachments created outside the resource graph. Unique within the
      # table: both engines key the association resource by this ID.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      attachment_id = string

      # Take over the attachment's EXISTING association instead of failing on
      # it. AWS allows one association per attachment gateway-wide, so
      # associating an attachment that is already associated somewhere else --
      # the gateway's default table, or another custom table -- is rejected at
      # apply time unless this flag is set, in which case the engine
      # disassociates the existing association first and associates here, in
      # one apply. Its primary use (per the provider's own guidance) is
      # attachments on a gateway SHARED INTO this account via RAM, where the
      # attachment's default-table membership is controlled by the sharing
      # account and cannot simply be turned off. The flag matters only when the
      # association is CREATED; changing it on an existing association is a
      # no-op. Leave it false for attachments whose default-table membership is
      # already off (the segmented-topology norm).
      replace_existing_association = optional(bool, false)
    })), [])

    # Attachments that PROPAGATE their routes into this table: their CIDRs
    # (VPC CIDRs for VPC attachments, BGP-learned routes for VPN and Direct
    # Connect) appear here automatically and track changes. An attachment can
    # propagate to many tables. Reference AwsTransitGatewayVpcAttachment
    # attachment_id outputs, or pass literal attachment IDs for attachments
    # created outside the resource graph.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    propagations = optional(list(string), [])

    # Static routes in this table. Statics win over propagated routes for the
    # same destination -- use them to steer specific prefixes through an
    # inspection attachment, provide a default route (0.0.0.0/0) toward an
    # egress VPC, or blackhole traffic that must never cross the hub.
    # Destination CIDRs must be unique within the table.
    routes = optional(list(object({
      # Destination CIDR block (IPv4 or IPv6), e.g. "10.20.0.0/16" or
      # "0.0.0.0/0" for a default route. Unique within the table. AWS uses
      # longest-prefix match across static and propagated routes, with statics
      # winning ties.
      destination_cidr_block = string

      # The attachment traffic to this destination is forwarded through.
      # Required unless blackhole is true. Reference an
      # AwsTransitGatewayVpcAttachment's attachment_id output, or pass a
      # literal attachment ID (VPN / Direct Connect / peering).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      attachment_id = optional(string, "")

      # Drop traffic to this destination instead of forwarding it. Mutually
      # exclusive with attachment_id. Blackholes are the segmentation kill
      # switch: a spoke's CIDR blackholed in another domain's table can never
      # be reached from that domain, regardless of propagations.
      blackhole = optional(bool, false)
    })), [])

    # Managed prefix list references. Each entry routes the entire CIDR set
    # of a customer-managed or AWS-managed prefix list via one attachment (or
    # blackholes it), tracking the list's membership as it changes --
    # operationally safer than mirroring a team's CIDR inventory as static
    # routes. Prefix list IDs must be unique within the table.
    prefix_list_references = optional(list(object({
      # The managed prefix list ID (e.g. "pl-0123456789abcdef0"). Unique within
      # the table. Both customer-managed and AWS-managed prefix lists are
      # accepted; the list's CIDR entries become routes that track membership
      # changes automatically.
      prefix_list_id = string

      # The attachment traffic to the prefix list's CIDRs is forwarded through.
      # Required unless blackhole is true. Reference an
      # AwsTransitGatewayVpcAttachment's attachment_id output, or pass a
      # literal attachment ID.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      attachment_id = optional(string, "")

      # Drop traffic to the prefix list's CIDRs instead of forwarding it.
      # Mutually exclusive with attachment_id.
      blackhole = optional(bool, false)
    })), [])

    # Designate this table as the gateway's DEFAULT ASSOCIATION route table:
    # attachments that keep default_route_table_association enabled (the
    # gateway-inherited posture) associate HERE instead of the AWS-created
    # default table. Meaningful only on a gateway whose own
    # default_route_table_association dial is enabled -- a gateway created
    # with the dial disabled has no default-association behavior to redirect.
    # At most one table per gateway may hold this designation; documents
    # cannot see each other, so AWS state -- not validation -- is the referee
    # and the most recent apply wins the pointer. Removing the flag RESTORES
    # the gateway's original default table (recorded at designation time by
    # both engines' providers).
    set_as_default_association_table = optional(bool, false)

    # Designate this table as the gateway's DEFAULT PROPAGATION route table:
    # attachments that keep default_route_table_propagation enabled advertise
    # their routes HERE instead of the AWS-created default table. Same
    # contract as set_as_default_association_table: meaningful only on a
    # gateway with the propagation dial enabled, at most one claimant per
    # gateway, and removing the flag restores the original table.
    set_as_default_propagation_table = optional(bool, false)
  })
}
