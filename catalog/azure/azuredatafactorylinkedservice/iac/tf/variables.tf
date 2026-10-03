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
  description = "AzureDataFactoryLinkedService specification"
  type = object({
    data_factory_id          = string
    name                     = string
    description              = optional(string, "")
    annotations              = optional(list(string), [])
    parameters               = optional(map(string), {})
    additional_properties    = optional(map(string), {})
    integration_runtime_name = optional(string, "")

    azure_blob_storage = optional(object({
      connection_string          = optional(string, "")
      connection_string_insecure = optional(string, "")
      sas_uri                    = optional(string, "")
      service_endpoint           = optional(string, "")
      sas_token_linked_key_vault_key = optional(object({
        linked_service_name = string
        secret_name         = string
      }))
      service_principal_linked_key_vault_key = optional(object({
        linked_service_name = string
        secret_name         = string
      }))
      storage_kind          = optional(string, "")
      use_managed_identity  = optional(bool)
      service_principal_id  = optional(string, "")
      service_principal_key = optional(string, "")
      tenant_id             = optional(string, "")
    }))

    azure_databricks = optional(object({
      adb_domain         = string
      msi_workspace_id   = optional(string, "")
      access_token       = optional(string, "")
      key_vault_password = optional(object({
        linked_service_name = string
        secret_name         = string
      }))
      existing_cluster_id = optional(string, "")
      new_cluster_config = optional(object({
        node_type                   = string
        cluster_version             = string
        min_number_of_workers       = optional(number)
        max_number_of_workers       = optional(number, 0)
        driver_node_type            = optional(string, "")
        log_destination             = optional(string, "")
        spark_config                = optional(map(string), {})
        spark_environment_variables = optional(map(string), {})
        custom_tags                 = optional(map(string), {})
        init_scripts                = optional(list(string), [])
      }))
      instance_pool = optional(object({
        instance_pool_id      = string
        cluster_version       = string
        min_number_of_workers = optional(number)
        max_number_of_workers = optional(number, 0)
      }))
    }))

    azure_file_storage = optional(object({
      connection_string = string
      file_share        = optional(string, "")
      host              = optional(string, "")
      user_id           = optional(string, "")
      password          = optional(string, "")
      key_vault_password = optional(object({
        linked_service_name = string
        secret_name         = string
      }))
    }))

    azure_function = optional(object({
      url = string
      key = optional(string, "")
      key_vault_key = optional(object({
        linked_service_name = string
        secret_name         = string
      }))
    }))

    azure_search = optional(object({
      url                = string
      search_service_key = string
    }))

    azure_sql_database = optional(object({
      connection_string = optional(string, "")
      key_vault_connection_string = optional(object({
        linked_service_name = string
        secret_name         = string
      }))
      key_vault_password = optional(object({
        linked_service_name = string
        secret_name         = string
      }))
      use_managed_identity  = optional(bool)
      service_principal_id  = optional(string, "")
      service_principal_key = optional(string, "")
      tenant_id             = optional(string, "")
      credential_name       = optional(string, "")
    }))

    azure_table_storage = optional(object({
      connection_string = string
    }))

    cosmosdb = optional(object({
      connection_string = optional(string, "")
      account_endpoint  = optional(string, "")
      account_key       = optional(string, "")
      database          = optional(string, "")
    }))

    cosmosdb_mongoapi = optional(object({
      connection_string              = string
      database                       = optional(string, "")
      server_version_is_32_or_higher = optional(bool)
    }))

    custom = optional(object({
      type                           = string
      type_properties_json           = string
      integration_runtime_parameters = optional(map(string), {})
    }))

    data_lake_storage_gen2 = optional(object({
      url                   = string
      use_managed_identity  = optional(bool)
      storage_account_key   = optional(string, "")
      service_principal_id  = optional(string, "")
      service_principal_key = optional(string, "")
      tenant                = optional(string, "")
    }))

    key_vault = optional(object({
      key_vault_id = string
    }))

    kusto = optional(object({
      kusto_endpoint        = string
      kusto_database_name   = string
      use_managed_identity  = optional(bool)
      service_principal_id  = optional(string, "")
      service_principal_key = optional(string, "")
      tenant                = optional(string, "")
    }))

    mysql = optional(object({
      connection_string = string
      driver_version    = optional(string)
    }))

    odata = optional(object({
      url = string
      basic_authentication = optional(object({
        username = string
        password = string
      }))
    }))

    odbc = optional(object({
      connection_string = string
      basic_authentication = optional(object({
        username = string
        password = string
      }))
    }))

    postgresql = optional(object({
      connection_string = string
    }))

    sftp = optional(object({
      authentication_type = string
      host                = string
      port                = number
      username            = string
      password            = optional(string, "")
      key_vault_password = optional(object({
        linked_service_name = string
        secret_name         = string
      }))
      private_key_content_base64 = optional(string, "")
      key_vault_private_key_content_base64 = optional(object({
        linked_service_name = string
        secret_name         = string
      }))
      private_key_path       = optional(string, "")
      private_key_passphrase = optional(string, "")
      key_vault_private_key_passphrase = optional(object({
        linked_service_name = string
        secret_name         = string
      }))
      skip_host_key_validation = optional(bool)
      host_key_fingerprint     = optional(string, "")
    }))

    snowflake = optional(object({
      connection_string = string
      key_vault_password = optional(object({
        linked_service_name = string
        secret_name         = string
      }))
    }))

    sql_managed_instance = optional(object({
      connection_string = optional(string, "")
      key_vault_connection_string = optional(object({
        linked_service_name = string
        secret_name         = string
      }))
      key_vault_password = optional(object({
        linked_service_name = string
        secret_name         = string
      }))
      service_principal_id  = optional(string, "")
      service_principal_key = optional(string, "")
      tenant                = optional(string, "")
    }))

    sql_server = optional(object({
      connection_string = optional(string, "")
      key_vault_connection_string = optional(object({
        linked_service_name = string
        secret_name         = string
      }))
      key_vault_password = optional(object({
        linked_service_name = string
        secret_name         = string
      }))
      user_name = optional(string, "")
    }))

    synapse = optional(object({
      connection_string = string
      key_vault_password = optional(object({
        linked_service_name = string
        secret_name         = string
      }))
    }))

    web = optional(object({
      url                 = string
      authentication_type = string
      username            = optional(string, "")
      password            = optional(string, "")
    }))
  })
}
