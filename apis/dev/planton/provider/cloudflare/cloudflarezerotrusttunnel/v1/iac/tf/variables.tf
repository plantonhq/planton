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
  description = "CloudflareZeroTrustTunnelSpec defines a Cloudflare Tunnel and its optional ingress config"
  type = object({
    # (Required) The Cloudflare account ID that owns this tunnel.
    account_id = string

    # (Required) A user-friendly name for the tunnel.
    name = string

    # (Optional) "cloudflare" (default, remote-managed) or "local" (origin YAML).
    config_src = optional(string, "cloudflare")

    # (Optional) Base64 secret for a locally-managed tunnel. Sensitive.
    tunnel_secret = optional(string, "")

    # (Optional) Public-hostname ingress rules; the last must be a catch-all.
    ingress = optional(list(object({
      hostname = optional(string, "")
      service  = string
      path     = optional(string, "")
      origin_request = optional(object({
        access = optional(object({
          # aud_tag is a repeated StringValueOrRef flattened to plain strings.
          aud_tag   = optional(list(string), [])
          team_name = optional(string, "")
          required  = optional(bool, false)
        }))
        ca_pool                  = optional(string, "")
        connect_timeout          = optional(number, 0)
        disable_chunked_encoding = optional(bool, false)
        http2_origin             = optional(bool, false)
        http_host_header         = optional(string, "")
        keep_alive_connections   = optional(number, 0)
        keep_alive_timeout       = optional(number, 0)
        match_sni_to_host        = optional(bool, false)
        no_happy_eyeballs        = optional(bool, false)
        no_tls_verify            = optional(bool, false)
        origin_server_name       = optional(string, "")
        proxy_type               = optional(string, "")
        tcp_keep_alive           = optional(number, 0)
        tls_timeout              = optional(number, 0)
      }))
    })), [])

    # (Optional) Tunnel-level origin defaults applied to every ingress rule.
    origin_request = optional(object({
      access = optional(object({
        aud_tag   = optional(list(string), [])
        team_name = optional(string, "")
        required  = optional(bool, false)
      }))
      ca_pool                  = optional(string, "")
      connect_timeout          = optional(number, 0)
      disable_chunked_encoding = optional(bool, false)
      http2_origin             = optional(bool, false)
      http_host_header         = optional(string, "")
      keep_alive_connections   = optional(number, 0)
      keep_alive_timeout       = optional(number, 0)
      match_sni_to_host        = optional(bool, false)
      no_happy_eyeballs        = optional(bool, false)
      no_tls_verify            = optional(bool, false)
      origin_server_name       = optional(string, "")
      proxy_type               = optional(string, "")
      tcp_keep_alive           = optional(number, 0)
      tls_timeout              = optional(number, 0)
    }))
  })
}
