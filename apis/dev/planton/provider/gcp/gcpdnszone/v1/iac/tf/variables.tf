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
  description = "spec"
  type = object({

    # The ID of the GCP project where the Managed Zone is created.
    # Supports StringValueOrRef pattern - use {value: "project-id"} for literal values.
    project_id = object({
      value = string
    })

    # An optional list of GCP service accounts that are granted permissions to manage DNS records in the Managed Zone.
    # These accounts are typically workload identities, such as those used by cert-manager,
    # and are added when new environments are created or updated.
    iam_service_accounts = optional(list(string), [])

    # The DNS records to be added to the Managed Zone.
    records = optional(list(object({

      # Required.** The DNS record type (e.g., A, AAAA, CNAME).
      record_type = string

      # Required.** The name of the DNS record (e.g., "example.com." or "dev.example.com.").
      # This value should always end with a dot to signify a fully qualified domain name.
      name = string

      # The list of values for the DNS record.
      # If the record type is CNAME, each value in the list should end with a dot.
      values = list(string)

      # The Time To Live (TTL) for the DNS record, in seconds.
      ttl_seconds = optional(number, 60)
    })), [])
  })
}
