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
  description = "AzureMachineLearningBatchDeployment specification"
  type = object({
    endpoint_id = string
    name        = string
    region      = string
    compute_id  = optional(string, "")
    resources = optional(object({
      instance_count = optional(number)
      instance_type  = optional(string, "")
    }))
    model = optional(object({
      id = optional(object({
        asset_id = string
      }))
      data_path = optional(object({
        datastore_id = optional(string, "")
        path         = optional(string, "")
      }))
      output_path = optional(object({
        job_id = optional(string, "")
        path   = optional(string, "")
      }))
    }))
    code_configuration = optional(object({
      code_id        = optional(string, "")
      scoring_script = string
    }))
    environment_id               = optional(string, "")
    environment_variables        = optional(map(string), {})
    mini_batch_size              = optional(number)
    max_concurrency_per_instance = optional(number)
    error_threshold              = optional(number)
    retry_settings = optional(object({
      max_retries = optional(number)
      timeout     = optional(string, "")
    }))
    output_action    = optional(string, "")
    output_file_name = optional(string, "")
    logging_level    = optional(string, "")
    pipeline_component = optional(object({
      component_id    = string
      settings        = optional(map(string), {})
      job_tags        = optional(map(string), {})
      job_description = optional(string, "")
    }))
    properties  = optional(map(string), {})
    description = optional(string, "")
    tags        = optional(map(string), {})
  })
}