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
  description = "CloudflareListItemSpec defines a single entry in a Cloudflare List"
  type = object({
    # (Required) The Cloudflare account ID that owns the parent list.
    account_id = string

    # (Required) The list ID. StringValueOrRef is flattened to a plain string
    # by the tfvars converter.
    list_id = optional(string)

    # Exactly one of the following matches the parent list's kind.
    ip  = optional(string)
    asn = optional(number)
    hostname = optional(object({
      url_hostname           = string
      exclude_exact_hostname = optional(bool)
    }))
    redirect = optional(object({
      source_url            = string
      target_url            = string
      status_code           = optional(number)
      include_subdomains    = optional(bool)
      preserve_path_suffix  = optional(bool)
      preserve_query_string = optional(bool)
      subpath_matching      = optional(bool)
    }))

    # (Optional) Informative summary of this entry.
    comment = optional(string, "")
  })
}
