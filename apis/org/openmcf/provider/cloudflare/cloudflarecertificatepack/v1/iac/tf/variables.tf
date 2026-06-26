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
  description = "CloudflareCertificatePackSpec defines an advanced certificate pack"
  type = object({
    # (Required) The Cloudflare Zone ID. StringValueOrRef is flattened to a plain
    # string by the tfvars converter.
    zone_id = optional(string)

    # (Required) "google", "lets_encrypt", or "ssl_com".
    certificate_authority = string

    # (Optional) Only "advanced" is supported (default).
    type = optional(string, "")

    # (Required) "txt", "http", or "email".
    validation_method = string

    # (Required) 14, 30, 90, or 365.
    validity_days = number

    # (Required) Hosts the certificate covers (must include the zone apex, ≤50).
    hosts = list(string)

    # (Optional) Add Cloudflare branding to the order.
    cloudflare_branding = optional(bool, false)
  })
}
