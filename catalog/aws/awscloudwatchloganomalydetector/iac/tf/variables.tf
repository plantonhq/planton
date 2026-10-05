variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsCloudwatchLogAnomalyDetector specification"
  type = object({
    region = string
    detector_name = optional(string, "")
    log_group_arns = list(string)
    enabled = optional(bool, false)
    evaluation_frequency = optional(string, "")
    filter_pattern = optional(string, "")
    kms_key_id = optional(string, "")
    anomaly_visibility_time = optional(number)
  })
}