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
  description = "CloudflareCustomHostnameSpec defines a Cloudflare for SaaS custom hostname"
  type = object({
    # (Required) The SaaS Zone ID. StringValueOrRef is flattened to a plain string.
    zone_id = optional(string)

    # (Required) The customer's hostname to onboard.
    hostname = string

    # (Optional) Override the origin for this hostname (StringValueOrRef -> string).
    custom_origin_server = optional(string, "")

    # (Optional) SNI sent to the custom origin.
    custom_origin_sni = optional(string, "")

    # (Optional) Arbitrary metadata stored with the hostname.
    custom_metadata = optional(map(string), {})

    # (Optional) SSL/TLS settings for the per-hostname certificate.
    ssl = optional(object({
      bundle_method         = optional(string, "")
      certificate_authority = optional(string, "")
      cloudflare_branding   = optional(bool, false)
      method                = optional(string, "")
      type                  = optional(string, "")
      wildcard              = optional(bool, false)
      custom_certificate    = optional(string, "")
      custom_csr_id         = optional(string, "")
      custom_key            = optional(string, "")
      custom_cert_bundle = optional(list(object({
        custom_certificate = string
        custom_key         = string
      })), [])
      settings = optional(object({
        ciphers         = optional(list(string), [])
        early_hints     = optional(string, "")
        http2           = optional(string, "")
        min_tls_version = optional(string, "")
        tls_1_3         = optional(string, "")
      }))
    }))
  })
}
