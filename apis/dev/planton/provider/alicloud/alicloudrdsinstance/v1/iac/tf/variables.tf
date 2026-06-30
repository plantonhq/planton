variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name   = string
    id     = optional(string)
    org    = optional(string)
    env    = optional(string)
    labels = optional(map(string))
    tags   = optional(list(string))
  })
}

variable "spec" {
  description = "Alibaba Cloud RDS Instance specification"
  type = object({
    region                    = string
    engine                    = string
    engine_version            = string
    instance_type             = string
    instance_storage          = number
    vswitch_id                = string
    instance_name             = optional(string, "")
    instance_charge_type      = optional(string, "Postpaid")
    category                  = optional(string, "HighAvailability")
    db_instance_storage_type  = optional(string, "")
    zone_id                   = optional(string, "")
    zone_id_slave_a           = optional(string, "")
    security_ips              = optional(list(string), [])
    security_group_ids        = optional(list(string), [])
    monitoring_period         = optional(number, null)
    maintain_time             = optional(string, "")
    deletion_protection       = optional(bool, false)
    ssl_action                = optional(string, "")
    tde_status                = optional(string, "")
    encryption_key            = optional(string, "")
    auto_renew                = optional(bool, false)
    auto_renew_period         = optional(number, null)
    period                    = optional(number, null)
    resource_group_id         = optional(string, "")
    tags                      = optional(map(string), {})
    parameters = optional(list(object({
      name  = string
      value = string
    })), [])
    databases = optional(list(object({
      name          = string
      character_set = optional(string, "")
      description   = optional(string, "")
    })), [])
    accounts = optional(list(object({
      account_name        = string
      account_password    = string
      account_type        = optional(string, "Normal")
      account_description = optional(string, "")
      privileges = optional(list(object({
        database_names = list(string)
        privilege      = optional(string, "ReadOnly")
      })), [])
    })), [])
  })

  validation {
    condition     = contains(["MySQL", "PostgreSQL", "SQLServer", "MariaDB", "PPAS"], var.spec.engine)
    error_message = "engine must be one of: MySQL, PostgreSQL, SQLServer, MariaDB, PPAS."
  }

  validation {
    condition     = contains(["Postpaid", "Prepaid"], var.spec.instance_charge_type)
    error_message = "instance_charge_type must be one of: Postpaid, Prepaid."
  }

  validation {
    condition     = contains(["Basic", "HighAvailability", "AlwaysOn", "Finance", "cluster"], var.spec.category)
    error_message = "category must be one of: Basic, HighAvailability, AlwaysOn, Finance, cluster."
  }

  validation {
    condition = (
      var.spec.db_instance_storage_type == "" ||
      contains(["local_ssd", "cloud_ssd", "cloud_essd", "cloud_essd2", "cloud_essd3"], var.spec.db_instance_storage_type)
    )
    error_message = "db_instance_storage_type must be one of: local_ssd, cloud_ssd, cloud_essd, cloud_essd2, cloud_essd3."
  }

  validation {
    condition = (
      var.spec.ssl_action == "" ||
      contains(["Open", "Close"], var.spec.ssl_action)
    )
    error_message = "ssl_action must be one of: Open, Close."
  }

  validation {
    condition = (
      var.spec.tde_status == "" ||
      contains(["Enabled", "Disabled"], var.spec.tde_status)
    )
    error_message = "tde_status must be one of: Enabled, Disabled."
  }
}
