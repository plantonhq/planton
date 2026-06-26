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
  description = "CloudflareEmailRoutingZoneSpec enables Email Routing on a zone"
  type = object({
    # (Required) The zone ID. StringValueOrRef is flattened to a plain string.
    zone_id = optional(string)

    # (Optional) The single per-zone catch-all rule.
    catch_all = optional(object({
      enabled    = optional(bool, false)
      type       = string
      forward_to = optional(list(string), [])
      worker     = optional(string, "")
    }))

    # (Optional) Lock the Email Routing DNS records.
    lock_dns_records = optional(bool, false)
  })
}
