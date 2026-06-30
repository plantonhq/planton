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
  description = "Specification for the DigitalOcean Database Cluster"
  type = object({
    cluster_name                = string
    engine                      = string  # "postgres", "mysql", "redis", "mongodb"
    engine_version              = string  # Major or major.minor version
    region                      = string  # DigitalOcean region
    size_slug                   = string  # Node size (e.g., "db-s-2vcpu-4gb")
    node_count                  = number  # 1-3 nodes
    vpc                         = optional(object({ value = string }))
    storage_gib                 = optional(number)
    enable_public_connectivity  = optional(bool, false)
  })
}
