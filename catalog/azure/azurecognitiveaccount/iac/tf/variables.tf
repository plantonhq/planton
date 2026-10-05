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
  description = "AzureCognitiveAccount specification"
  type = object({
    region                     = string
    resource_group             = string
    name                       = string
    kind                       = string
    sku_name                   = string
    project_management_enabled = optional(bool, false)
    custom_subdomain_name      = optional(string, "")
    customer_managed_key = optional(object({
      key_vault_key_id   = string
      identity_client_id = optional(string, "")
    }))
    dynamic_throttling_enabled = optional(bool, false)
    fqdns                      = optional(list(string), [])
    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))
    local_auth_enabled              = optional(bool)
    metrics_advisor_aad_client_id   = optional(string, "")
    metrics_advisor_aad_tenant_id   = optional(string, "")
    metrics_advisor_super_user_name = optional(string, "")
    metrics_advisor_website_name    = optional(string, "")
    network_acls = optional(object({
      default_action = string
      ip_rules       = optional(list(string), [])
      virtual_network_rules = optional(list(object({
        subnet_id                            = string
        ignore_missing_vnet_service_endpoint = optional(bool, false)
      })), [])
      bypass = optional(string, "")
    }))
    network_injection = optional(object({
      scenario  = string
      subnet_id = string
    }))
    outbound_network_access_restricted           = optional(bool, false)
    public_network_access_enabled                = optional(bool)
    qna_runtime_endpoint                         = optional(string, "")
    custom_question_answering_search_service_id  = optional(string, "")
    custom_question_answering_search_service_key = optional(string, "")
    storage = optional(list(object({
      storage_account_id = string
      identity_client_id = optional(string, "")
    })), [])
    tags = optional(map(string), {})
    rai_blocklists = optional(list(object({
      name        = string
      description = optional(string, "")
      tags        = optional(map(string), {})
    })), [])
    rai_policies = optional(list(object({
      name             = string
      base_policy_name = string
      content_filters = list(object({
        name               = string
        filter_enabled     = optional(bool, false)
        block_enabled      = optional(bool, false)
        source             = string
        severity_threshold = optional(string, "")
      }))
      mode = optional(string, "")
      tags = optional(map(string), {})
    })), [])
  })
}