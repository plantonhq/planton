variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "OciRedisCluster specification"
  type = object({
    compartment_id = object({
      value = string
    })

    display_name = optional(string, "")

    subnet_id = object({
      value = string
    })

    node_count        = number
    node_memory_in_gbs = number
    software_version  = string

    cluster_mode = optional(string, "")
    shard_count  = optional(number, 0)

    nsg_ids = optional(list(object({
      value = string
    })), [])

    config_set_id = optional(object({
      value = string
    }), null)
  })
}
