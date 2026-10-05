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
  description = "CloudflareLogpushJob specification"
  type = object({
    account_id = optional(string, "")
    zone_id = optional(string, "")
    dataset = string
    destination_conf = string
    name = optional(string, "")
    enabled = optional(bool)
    filter = optional(string, "")
    kind = optional(string, "")
    max_upload_bytes = optional(number)
    max_upload_interval_seconds = optional(number)
    max_upload_records = optional(number)
    output_options = optional(object({
      output_type = optional(string, "")
      field_names = optional(list(string), [])
      timestamp_format = optional(string, "")
      sample_rate = optional(number)
      batch_prefix = optional(string, "")
      batch_suffix = optional(string, "")
      record_prefix = optional(string, "")
      record_suffix = optional(string, "")
      record_delimiter = optional(string, "")
      record_template = optional(string, "")
      field_delimiter = optional(string, "")
      merge_subrequests = optional(bool)
      cve_2021_44228 = optional(bool)
    }))
    ownership_challenge = optional(string, "")
    generate_ownership_challenge = optional(bool, false)
  })
}