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
  description = "AzureDataFactoryIntegrationRuntime specification"
  type = object({
    data_factory_id = string
    name            = string
    description     = optional(string, "")

    azure = optional(object({
      region                                        = string
      cleanup_enabled                               = optional(bool)
      compute_type                                  = optional(string, "")
      core_count                                    = optional(number, 0)
      time_to_live_min                              = optional(number, 0)
      interactive_authoring_time_to_live_in_minutes = optional(number, 0)
      virtual_network_enabled                       = optional(bool, false)
    }))

    azure_ssis = optional(object({
      region                           = string
      node_size                        = string
      number_of_nodes                  = optional(number, 0)
      max_parallel_executions_per_node = optional(number, 0)
      edition                          = optional(string, "")
      license_type                     = optional(string, "")
      credential_name                  = optional(string, "")

      catalog_info = optional(object({
        server_endpoint        = string
        administrator_login    = optional(string, "")
        administrator_password = optional(string, "")
        pricing_tier           = optional(string, "")
        elastic_pool_name      = optional(string, "")
        dual_standby_pair_name = optional(string, "")
      }))

      custom_setup_script = optional(object({
        blob_container_uri = string
        sas_token          = string
      }))

      express_custom_setup = optional(object({
        environment        = optional(map(string), {})
        powershell_version = optional(string, "")
        command_key = optional(list(object({
          target_name = string
          user_name   = string
          password    = optional(string, "")
          key_vault_password = optional(object({
            linked_service_name = string
            secret_name         = string
            parameters          = optional(map(string), {})
            secret_version      = optional(string, "")
          }))
        })), [])
        component = optional(list(object({
          name    = string
          license = optional(string, "")
          key_vault_license = optional(object({
            linked_service_name = string
            secret_name         = string
            parameters          = optional(map(string), {})
            secret_version      = optional(string, "")
          }))
        })), [])
      }))

      express_vnet_integration = optional(object({
        subnet_id = string
      }))

      vnet_integration = optional(object({
        vnet_id     = optional(string, "")
        subnet_id   = optional(string, "")
        subnet_name = optional(string, "")
        public_ips  = optional(list(string), [])
      }))

      package_store = optional(list(object({
        name                = string
        linked_service_name = string
      })), [])

      copy_compute_scale = optional(object({
        data_integration_unit = optional(number, 0)
        time_to_live          = optional(number, 0)
      }))

      pipeline_external_compute_scale = optional(object({
        number_of_external_nodes = optional(number, 0)
        number_of_pipeline_nodes = optional(number, 0)
        time_to_live             = optional(number, 0)
      }))

      proxy = optional(object({
        self_hosted_integration_runtime_name = string
        staging_storage_linked_service_name  = string
        path                                 = optional(string, "")
      }))
    }))

    self_hosted = optional(object({
      rbac_authorization = optional(object({
        resource_id = string
      }))
      self_contained_interactive_authoring_enabled = optional(bool, false)
    }))
  })
}
