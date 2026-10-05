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
  description = "AzureDataFactoryPipeline specification"
  type = object({
    data_factory_id                = string
    name                           = string
    description                    = optional(string, "")
    parameters                     = optional(map(string), {})
    variables                      = optional(map(string), {})
    activities_json                = optional(string, "")
    annotations                    = optional(list(string), [])
    concurrency                    = optional(number)
    folder                         = optional(string, "")
    monitor_metrics_after_duration = optional(string, "")
  })
}
