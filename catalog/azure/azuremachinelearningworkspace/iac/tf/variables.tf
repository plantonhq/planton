variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "AzureMachineLearningWorkspace specification"
  type = object({
    region                  = string
    resource_group          = string
    name                    = string
    application_insights_id = string
    key_vault_id            = string
    storage_account_id      = string
    identity = object({
      type         = string
      identity_ids = optional(list(string), [])
    })
    kind = optional(string, "")
    feature_store = optional(object({
      enabled                        = optional(bool)
      computer_spark_runtime_version = optional(string, "")
      offline_connection_name        = optional(string, "")
      online_connection_name         = optional(string, "")
    }))
    primary_user_assigned_identity = optional(string, "")
    container_registry_id          = optional(string, "")
    public_network_access_enabled  = optional(bool)
    image_build_compute_name       = optional(string, "")
    description                    = optional(string, "")
    friendly_name                  = optional(string, "")
    encryption = optional(object({
      key_vault_id              = string
      key_id                    = string
      user_assigned_identity_id = optional(string, "")
    }))
    managed_network = optional(object({
      isolation_mode                = optional(string, "")
      provision_on_creation_enabled = optional(bool, false)
    }))
    high_business_impact            = optional(bool, false)
    sku_name                        = optional(string, "")
    service_side_encryption_enabled = optional(bool, false)
    v1_legacy_mode_enabled          = optional(bool, false)
    storage_account_access_type     = optional(string, "")
    serverless_compute = optional(object({
      subnet_id         = optional(string, "")
      public_ip_enabled = optional(bool, false)
    }))
    tags = optional(map(string), {})
    fqdn_outbound_rules = optional(list(object({
      name             = string
      destination_fqdn = string
    })), [])
    private_endpoint_outbound_rules = optional(list(object({
      name                = string
      service_resource_id = string
      sub_resource_target = string
      spark_enabled       = optional(bool, false)
    })), [])
    service_tag_outbound_rules = optional(list(object({
      name        = string
      service_tag = string
      protocol    = string
      port_ranges = string
    })), [])
  })
}