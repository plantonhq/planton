variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name = string,
    id = optional(string),
    org = optional(string),
    env = optional(string),
    labels = optional(map(string)),
    tags = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "CloudflareD1DatabaseSpec defines the essential configuration for creating a Cloudflare D1 database"
  type = object({
    # (Required) The Cloudflare account ID in which to create the database.
    account_id = string

    # (Required) The unique name for the D1 database.
    # Must be unique within the account.
    database_name = string

    # (Optional) The Cloudflare region where the D1 database will be hosted.
    # Valid values: "weur", "eeur", "apac", "oc", "wnam", "enam".
    # If omitted, Cloudflare selects a default location based on your account settings.
    region = optional(string)

    # (Optional) Configures D1 Read Replication.
    # Enables automatic read replication across multiple regions for lower global read latency.
    # WARNING: Enabling replication requires application-level code changes to use the D1 Sessions API.
    read_replication = optional(object({
      # (Required if read_replication is set) The replication mode.
      # Valid values: "auto" (enable automatic read replication), "disabled" (disable replication).
      mode = string
    }))
  })
}