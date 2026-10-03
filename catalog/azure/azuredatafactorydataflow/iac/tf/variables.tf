variable "metadata" {
  description = "Catalog object metadata"
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
  description = "AzureDataFactoryDataFlow specification"
  type = object({
    data_factory_id = string
    name            = string
    flowlet         = optional(bool, false)
    script          = optional(string, "")
    script_lines    = optional(list(string), [])
    sources = optional(list(object({
      name        = string
      description = optional(string, "")
      dataset = optional(object({
        name       = string
        parameters = optional(map(string), {})
      }))
      flowlet = optional(object({
        name               = string
        parameters         = optional(map(string), {})
        dataset_parameters = optional(string, "")
      }))
      linked_service = optional(object({
        name       = string
        parameters = optional(map(string), {})
      }))
      schema_linked_service = optional(object({
        name       = string
        parameters = optional(map(string), {})
      }))
    })), [])
    sinks = optional(list(object({
      name        = string
      description = optional(string, "")
      dataset = optional(object({
        name       = string
        parameters = optional(map(string), {})
      }))
      flowlet = optional(object({
        name               = string
        parameters         = optional(map(string), {})
        dataset_parameters = optional(string, "")
      }))
      linked_service = optional(object({
        name       = string
        parameters = optional(map(string), {})
      }))
      schema_linked_service = optional(object({
        name       = string
        parameters = optional(map(string), {})
      }))
      rejected_linked_service = optional(object({
        name       = string
        parameters = optional(map(string), {})
      }))
    })), [])
    transformations = optional(list(object({
      name        = string
      description = optional(string, "")
      dataset = optional(object({
        name       = string
        parameters = optional(map(string), {})
      }))
      flowlet = optional(object({
        name               = string
        parameters         = optional(map(string), {})
        dataset_parameters = optional(string, "")
      }))
      linked_service = optional(object({
        name       = string
        parameters = optional(map(string), {})
      }))
    })), [])
    description = optional(string, "")
    annotations = optional(list(string), [])
    folder      = optional(string, "")
  })
}