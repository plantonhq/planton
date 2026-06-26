variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "CloudflareZeroTrustTunnelRouteSpec defines a Cloudflare Tunnel route"
  type = object({
    # (Required) The Cloudflare account ID that owns this route.
    account_id = string

    # (Required) The private CIDR advertised by this route.
    network = string

    # (Required) Tunnel that serves this network. StringValueOrRef flattened to a
    # plain tunnel UUID by the tfvars converter.
    tunnel_id = string

    # (Optional) Virtual network this route belongs to (UUID). Omit to use the account
    # default. StringValueOrRef flattened to a plain UUID by the tfvars converter.
    virtual_network_id = optional(string, "")

    # (Optional) Remark describing the route.
    comment = optional(string, "")
  })
}
