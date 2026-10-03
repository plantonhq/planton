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
  description = "AwsBackupReportPlan specification"
  type = object({
    region = string
    report_plan_name = string
    description = optional(string, "")
    delivery_channel = object({
      s3_bucket_name = string
      s3_key_prefix = optional(string, "")
      formats = optional(list(string), [])
    })
    report_setting = object({
      report_template = optional(string, "")
      framework_arns = optional(list(string), [])
      number_of_frameworks = optional(number, 0)
      accounts = optional(list(string), [])
      organization_units = optional(list(string), [])
      regions = optional(list(string), [])
    })
  })
}