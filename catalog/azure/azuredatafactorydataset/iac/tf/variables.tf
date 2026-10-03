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
  description = "AzureDataFactoryDataset specification"
  type = object({
    data_factory_id       = string
    name                  = string
    linked_service_name   = optional(string, "")
    description           = optional(string, "")
    annotations           = optional(list(string), [])
    parameters            = optional(map(string), {})
    additional_properties = optional(map(string), {})
    folder                = optional(string, "")

    azure_blob = optional(object({
      path                     = optional(string, "")
      filename                 = optional(string, "")
      dynamic_path_enabled     = optional(bool, false)
      dynamic_filename_enabled = optional(bool, false)
      schema_column = optional(list(object({
        name        = string
        type        = optional(string, "")
        description = optional(string, "")
      })), [])
    }))

    azure_sql_table = optional(object({
      linked_service_id = string
      schema            = optional(string, "")
      table             = optional(string, "")
      schema_column = optional(list(object({
        name        = string
        type        = optional(string, "")
        description = optional(string, "")
      })), [])
    }))

    binary = optional(object({
      http_server_location = optional(object({
        relative_url             = string
        path                     = optional(string, "")
        dynamic_path_enabled     = optional(bool, false)
        filename                 = optional(string, "")
        dynamic_filename_enabled = optional(bool, false)
      }))
      azure_blob_storage_location = optional(object({
        container                 = string
        dynamic_container_enabled = optional(bool, false)
        path                      = optional(string, "")
        dynamic_path_enabled      = optional(bool, false)
        filename                  = optional(string, "")
        dynamic_filename_enabled  = optional(bool, false)
      }))
      sftp_server_location = optional(object({
        path                     = string
        dynamic_path_enabled     = optional(bool, false)
        filename                 = string
        dynamic_filename_enabled = optional(bool, false)
      }))
      compression = optional(object({
        type  = string
        level = optional(string, "")
      }))
    }))

    cosmosdb_sqlapi = optional(object({
      collection_name = optional(string, "")
      schema_column = optional(list(object({
        name        = string
        type        = optional(string, "")
        description = optional(string, "")
      })), [])
    }))

    custom = optional(object({
      linked_service = object({
        name       = string
        parameters = optional(map(string), {})
      })
      type                 = string
      type_properties_json = string
      schema_json          = optional(string, "")
    }))

    delimited_text = optional(object({
      http_server_location = optional(object({
        relative_url             = string
        path                     = optional(string, "")
        dynamic_path_enabled     = optional(bool, false)
        filename                 = optional(string, "")
        dynamic_filename_enabled = optional(bool, false)
      }))
      azure_blob_storage_location = optional(object({
        container                 = string
        dynamic_container_enabled = optional(bool, false)
        path                      = optional(string, "")
        dynamic_path_enabled      = optional(bool, false)
        filename                  = optional(string, "")
        dynamic_filename_enabled  = optional(bool, false)
      }))
      azure_blob_fs_location = optional(object({
        file_system                 = optional(string, "")
        dynamic_file_system_enabled = optional(bool, false)
        path                        = optional(string, "")
        dynamic_path_enabled        = optional(bool, false)
        filename                    = optional(string, "")
        dynamic_filename_enabled    = optional(bool, false)
      }))
      column_delimiter    = optional(string, "")
      row_delimiter       = optional(string, "")
      quote_character     = optional(string, "")
      escape_character    = optional(string, "")
      encoding            = optional(string, "")
      first_row_as_header = optional(bool)
      null_value          = optional(string, "")
      compression_codec   = optional(string, "")
      compression_level   = optional(string, "")
      schema_column = optional(list(object({
        name        = string
        type        = optional(string, "")
        description = optional(string, "")
      })), [])
    }))

    http = optional(object({
      relative_url   = optional(string, "")
      request_body   = optional(string, "")
      request_method = optional(string, "")
      schema_column = optional(list(object({
        name        = string
        type        = optional(string, "")
        description = optional(string, "")
      })), [])
    }))

    json = optional(object({
      http_server_location = optional(object({
        relative_url             = string
        path                     = optional(string, "")
        dynamic_path_enabled     = optional(bool, false)
        filename                 = optional(string, "")
        dynamic_filename_enabled = optional(bool, false)
      }))
      azure_blob_storage_location = optional(object({
        container                 = string
        dynamic_container_enabled = optional(bool, false)
        path                      = optional(string, "")
        dynamic_path_enabled      = optional(bool, false)
        filename                  = optional(string, "")
        dynamic_filename_enabled  = optional(bool, false)
      }))
      encoding = optional(string, "")
      schema_column = optional(list(object({
        name        = string
        type        = optional(string, "")
        description = optional(string, "")
      })), [])
    }))

    mysql = optional(object({
      table_name = optional(string, "")
      schema_column = optional(list(object({
        name        = string
        type        = optional(string, "")
        description = optional(string, "")
      })), [])
    }))

    parquet = optional(object({
      http_server_location = optional(object({
        relative_url             = string
        path                     = optional(string, "")
        dynamic_path_enabled     = optional(bool, false)
        filename                 = optional(string, "")
        dynamic_filename_enabled = optional(bool, false)
      }))
      azure_blob_storage_location = optional(object({
        container                 = string
        dynamic_container_enabled = optional(bool, false)
        path                      = optional(string, "")
        dynamic_path_enabled      = optional(bool, false)
        filename                  = optional(string, "")
        dynamic_filename_enabled  = optional(bool, false)
      }))
      azure_blob_fs_location = optional(object({
        file_system                 = optional(string, "")
        dynamic_file_system_enabled = optional(bool, false)
        path                        = optional(string, "")
        dynamic_path_enabled        = optional(bool, false)
        filename                    = optional(string, "")
        dynamic_filename_enabled    = optional(bool, false)
      }))
      compression_codec = optional(string, "")
      schema_column = optional(list(object({
        name        = string
        type        = optional(string, "")
        description = optional(string, "")
      })), [])
    }))

    postgresql = optional(object({
      table_name = optional(string, "")
      schema_column = optional(list(object({
        name        = string
        type        = optional(string, "")
        description = optional(string, "")
      })), [])
    }))

    snowflake = optional(object({
      table_name  = optional(string, "")
      schema_name = optional(string, "")
      schema_column = optional(list(object({
        name      = string
        type      = optional(string, "")
        precision = optional(number, 0)
        scale     = optional(number, 0)
      })), [])
    }))

    sql_server_table = optional(object({
      table_name = optional(string, "")
      schema_column = optional(list(object({
        name        = string
        type        = optional(string, "")
        description = optional(string, "")
      })), [])
    }))
  })
}
