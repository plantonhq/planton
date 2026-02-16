variable "spec" {
  description = "GcpBigQueryDataset spec"
  type = object({
    project_id = object({ value = string })
    dataset_id = string
    location   = string

    friendly_name                  = optional(string, "")
    description                    = optional(string, "")
    default_table_expiration_ms    = optional(number, 0)
    default_partition_expiration_ms = optional(number, 0)
    max_time_travel_hours          = optional(number, 0)
    is_case_insensitive            = optional(bool, false)
    default_collation              = optional(string, "")
    storage_billing_model          = optional(string, "")
    delete_contents_on_destroy     = optional(bool, false)

    kms_key_name = optional(object({ value = string }), null)

    access = optional(list(object({
      role           = optional(string, "")
      user_by_email  = optional(string, "")
      group_by_email = optional(string, "")
      domain         = optional(string, "")
      special_group  = optional(string, "")
      iam_member     = optional(string, "")
      view = optional(object({
        project_id = string
        dataset_id = string
        table_id   = string
      }), null)
    })), [])
  })
}

variable "provider_config" {
  description = "GCP provider configuration"
  type = object({
    service_account_key_base64 = optional(string, "")
  })
  default = { service_account_key_base64 = "" }
}
