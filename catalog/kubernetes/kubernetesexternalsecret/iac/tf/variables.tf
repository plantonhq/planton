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
  description = "KubernetesExternalSecret specification"
  type = object({
    namespace = string
    store_ref = object({
      secret_store = optional(object({
        name = string
      }))
      cluster_secret_store = optional(object({
        name = string
      }))
    })
    refresh_interval = optional(string)
    refresh_policy   = optional(string)
    target = optional(object({
      name            = optional(string, "")
      creation_policy = optional(string)
      deletion_policy = optional(string)
      immutable       = optional(bool, false)
      template = optional(object({
        type         = optional(string, "")
        merge_policy = optional(string)
        labels       = optional(map(string), {})
        annotations  = optional(map(string), {})
        data         = optional(map(string), {})
      }))
    }))
    data = optional(list(object({
      secret_key = string
      remote_ref = object({
        key               = string
        property          = optional(string, "")
        version           = optional(string, "")
        decoding_strategy = optional(string)
      })
    })), [])
    data_from = optional(list(object({
      extract = optional(object({
        key               = string
        property          = optional(string, "")
        version           = optional(string, "")
        decoding_strategy = optional(string)
      }))
      find = optional(object({
        path        = optional(string, "")
        name_regexp = optional(string, "")
        tags        = optional(map(string), {})
      }))
      rewrite = optional(list(object({
        source = string
        target = string
      })), [])
    })), [])
  })
}
