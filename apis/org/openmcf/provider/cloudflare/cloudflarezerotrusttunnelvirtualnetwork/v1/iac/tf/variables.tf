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
  description = "CloudflareZeroTrustTunnelVirtualNetworkSpec defines a Cloudflare Tunnel virtual network"
  type = object({
    # (Required) The Cloudflare account ID that owns this virtual network.
    account_id = string

    # (Required) A user-friendly name for the virtual network.
    name = string

    # (Optional) Remark describing the virtual network's purpose.
    comment = optional(string, "")

    # (Optional) When true, this virtual network becomes the account default.
    is_default_network = optional(bool, false)
  })
}
