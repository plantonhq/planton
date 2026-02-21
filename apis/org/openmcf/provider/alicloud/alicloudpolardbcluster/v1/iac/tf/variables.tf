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
  description = "Alibaba Cloud PolarDB Cluster specification"
  type = object({
    region                                      = string
    db_type                                     = string
    db_version                                  = string
    db_node_class                               = string
    vswitch_id                                  = string
    db_node_count                               = optional(number, 2)
    description                                 = optional(string, "")
    pay_type                                    = optional(string, "PostPaid")
    period                                      = optional(number, null)
    renewal_status                              = optional(string, "")
    auto_renew_period                           = optional(number, null)
    zone_id                                     = optional(string, "")
    security_ips                                = optional(list(string), [])
    security_group_ids                          = optional(list(string), [])
    maintain_time                               = optional(string, "")
    resource_group_id                           = optional(string, "")
    tags                                        = optional(map(string), {})
    creation_category                           = optional(string, "")
    sub_category                                = optional(string, "")
    storage_type                                = optional(string, "")
    storage_space                               = optional(number, null)
    tde_status                                  = optional(string, "")
    encryption_key                              = optional(string, "")
    deletion_lock                               = optional(number, null)
    collector_status                            = optional(string, "")
    backup_retention_policy_on_cluster_deletion = optional(string, "")
    parameters = optional(list(object({
      name  = string
      value = string
    })), [])
    databases = optional(list(object({
      db_name            = string
      character_set_name = optional(string, "")
      db_description     = optional(string, "")
      collate            = optional(string, "")
      ctype              = optional(string, "")
    })), [])
    accounts = optional(list(object({
      account_name        = string
      account_password    = string
      account_type        = optional(string, "Normal")
      account_description = optional(string, "")
      privileges = optional(list(object({
        db_names          = list(string)
        account_privilege = optional(string, "ReadOnly")
      })), [])
    })), [])
  })

  validation {
    condition     = contains(["MySQL", "PostgreSQL", "Oracle"], var.spec.db_type)
    error_message = "db_type must be one of: MySQL, PostgreSQL, Oracle."
  }

  validation {
    condition     = contains(["PostPaid", "PrePaid"], var.spec.pay_type)
    error_message = "pay_type must be one of: PostPaid, PrePaid."
  }

  validation {
    condition = (
      var.spec.creation_category == "" ||
      contains(["Normal", "Basic", "ArchiveNormal", "NormalMultimaster", "SENormal"], var.spec.creation_category)
    )
    error_message = "creation_category must be one of: Normal, Basic, ArchiveNormal, NormalMultimaster, SENormal."
  }

  validation {
    condition = (
      var.spec.storage_type == "" ||
      contains(["PSL5", "PSL4", "ESSDPL0", "ESSDPL1", "ESSDPL2", "ESSDPL3", "ESSDAUTOPL"], var.spec.storage_type)
    )
    error_message = "storage_type must be one of: PSL5, PSL4, ESSDPL0, ESSDPL1, ESSDPL2, ESSDPL3, ESSDAUTOPL."
  }

  validation {
    condition = (
      var.spec.tde_status == "" ||
      contains(["Enabled", "Disabled"], var.spec.tde_status)
    )
    error_message = "tde_status must be one of: Enabled, Disabled."
  }
}
