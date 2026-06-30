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
  description = "CloudflareWorkersKvPairSpec defines a single key-value entry in a Workers KV namespace"
  type = object({
    # (Required) The Cloudflare account ID that owns the namespace.
    account_id = string

    # (Required) The KV namespace ID. StringValueOrRef is flattened to a plain
    # string by the tfvars converter.
    namespace_id = optional(string)

    # (Required) The entry key (up to 512 bytes).
    key_name = string

    # (Required) The value stored at key_name (up to 25 MiB).
    value = string

    # (Optional) Arbitrary JSON metadata associated with the entry.
    metadata = optional(string, "")
  })
}
