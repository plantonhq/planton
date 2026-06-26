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
  description = "CloudflareOriginCaCertificateSpec defines an Origin CA certificate"
  type = object({
    # (Required) Hostnames (SANs) the certificate is valid for (at least one).
    hostnames = list(string)

    # (Optional) "origin-rsa" (default), "origin-ecc", or "keyless-certificate".
    request_type = optional(string, "")

    # (Optional) Validity in days: 7, 30, 90, 365, 730, 1095, or 5475 (default).
    requested_validity = optional(number, 0)

    # (Optional) A user-supplied CSR (PEM). When set, no key is generated.
    csr = optional(string, "")
  })
}
