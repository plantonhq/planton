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
  description = "AwsBackupRestoreTestingPlan specification"
  type = object({
    region = string
    plan_name = string
    schedule_expression = string
    schedule_expression_timezone = optional(string, "")
    start_window_hours = optional(number, 0)
    recovery_point_selection = object({
      algorithm = optional(string, "")
      include_vaults = list(string)
      recovery_point_types = list(string)
      exclude_vaults = optional(list(string), [])
      selection_window_days = optional(number, 0)
    })
    selections = optional(list(object({
      name = string
      protected_resource_type = string
      iam_role_arn = string
      protected_resource_arns = optional(list(string), [])
      protected_resource_conditions = optional(object({
        string_equals = optional(list(object({
          key = string
          value = string
        })), [])
        string_not_equals = optional(list(object({
          key = string
          value = string
        })), [])
      }))
      restore_metadata_overrides = optional(map(string), {})
      validation_window_hours = optional(number, 0)
    })), [])
  })
}