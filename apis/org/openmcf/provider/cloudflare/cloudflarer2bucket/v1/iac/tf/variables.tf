variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for Cloudflare R2 Bucket"
  type = object({
    # DNS-compatible bucket name (3-63 characters, lowercase alphanumeric + hyphens)
    bucket_name = string

    # Cloudflare account ID (32 hex characters)
    account_id = string

    # Primary region for the bucket (location hint), e.g. "auto", "wnam",
    # "enam", "weur", "eeur", "apac", "oc". "auto" lets Cloudflare choose.
    location = optional(string)

    # Expose bucket via the managed r2.dev public URL.
    public_access = optional(bool, false)

    # Custom domain configuration for the bucket
    custom_domain = optional(object({
      # Whether to enable custom domain access for the bucket
      enabled = bool

      # The Cloudflare Zone ID hosting the custom domain (resolved literal).
      zone_id = optional(string)

      # The full domain name to use for accessing the bucket
      domain = string
    }))
  })
}
