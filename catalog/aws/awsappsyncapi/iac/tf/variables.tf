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
  description = "AwsAppSyncApi specification"
  type = object({
    region = string
    graphql = optional(object({
      api_name = string
      auth = object({
        type = optional(string, "")
        user_pool = optional(object({
          user_pool_id = optional(string, "")
          app_id_client_regex = optional(string, "")
          aws_region = optional(string, "")
          default_action = optional(string, "")
        }))
        openid_connect = optional(object({
          issuer = string
          client_id = optional(string, "")
          iat_ttl = optional(number, 0)
          auth_ttl = optional(number, 0)
        }))
        lambda = optional(object({
          authorizer_uri = optional(string, "")
          authorizer_result_ttl_in_seconds = optional(number, 0)
          identity_validation_expression = optional(string, "")
        }))
      })
      additional_auth_providers = optional(list(object({
        type = optional(string, "")
        user_pool = optional(object({
          user_pool_id = optional(string, "")
          app_id_client_regex = optional(string, "")
          aws_region = optional(string, "")
          default_action = optional(string, "")
        }))
        openid_connect = optional(object({
          issuer = string
          client_id = optional(string, "")
          iat_ttl = optional(number, 0)
          auth_ttl = optional(number, 0)
        }))
        lambda = optional(object({
          authorizer_uri = optional(string, "")
          authorizer_result_ttl_in_seconds = optional(number, 0)
          identity_validation_expression = optional(string, "")
        }))
      })), [])
      schema = optional(string, "")
      visibility = optional(string, "")
      disable_introspection = optional(bool, false)
      query_depth_limit = optional(number, 0)
      resolver_count_limit = optional(number, 0)
      xray_enabled = optional(bool, false)
      log_config = optional(object({
        cloudwatch_logs_role_arn = optional(string, "")
        field_log_level = optional(string, "")
        exclude_verbose_content = optional(bool, false)
      }))
      enhanced_metrics = optional(object({
        data_source_level_metrics_behavior = optional(string, "")
        operation_level_metrics_config = optional(string, "")
        resolver_level_metrics_behavior = optional(string, "")
      }))
      web_acl_arn = optional(string, "")
      cache = optional(object({
        api_caching_behavior = optional(string, "")
        ttl = optional(number, 0)
        type = optional(string, "")
        at_rest_encryption_enabled = optional(bool, false)
        transit_encryption_enabled = optional(bool, false)
      }))
      types = optional(list(object({
        name = string
        definition = string
        format = optional(string, "")
      })), [])
      functions = optional(list(object({
        name = string
        data_source_name = string
        description = optional(string, "")
        code = optional(string, "")
        runtime_version = optional(string, "")
        request_mapping_template = optional(string, "")
        response_mapping_template = optional(string, "")
        max_batch_size = optional(number, 0)
        sync_config = optional(object({
          conflict_detection = optional(string, "")
          conflict_handler = optional(string, "")
          lambda_conflict_handler_arn = optional(string, "")
        }))
      })), [])
      resolvers = optional(list(object({
        type = string
        field = string
        data_source_name = optional(string, "")
        pipeline_functions = optional(list(string), [])
        code = optional(string, "")
        runtime_version = optional(string, "")
        request_template = optional(string, "")
        response_template = optional(string, "")
        max_batch_size = optional(number, 0)
        caching = optional(object({
          caching_keys = optional(list(string), [])
          ttl = optional(number, 0)
        }))
        sync_config = optional(object({
          conflict_detection = optional(string, "")
          conflict_handler = optional(string, "")
          lambda_conflict_handler_arn = optional(string, "")
        }))
      })), [])
      merged = optional(object({
        execution_role_arn = string
        source_apis = optional(list(object({
          name = string
          source_api_id = string
          merge_type = optional(string, "")
          description = optional(string, "")
        })), [])
      }))
    }))
    events = optional(object({
      owner_contact = optional(string, "")
      auth_providers = list(object({
        type = optional(string, "")
        cognito = optional(object({
          user_pool_id = optional(string, "")
          aws_region = string
          app_id_client_regex = optional(string, "")
        }))
        openid_connect = optional(object({
          issuer = string
          client_id = optional(string, "")
          iat_ttl = optional(number, 0)
          auth_ttl = optional(number, 0)
        }))
        lambda = optional(object({
          authorizer_uri = optional(string, "")
          authorizer_result_ttl_in_seconds = optional(number, 0)
          identity_validation_expression = optional(string, "")
        }))
      }))
      connection_auth_modes = list(string)
      default_publish_auth_modes = list(string)
      default_subscribe_auth_modes = list(string)
      log_config = optional(object({
        cloudwatch_logs_role_arn = string
        log_level = optional(string, "")
      }))
      channel_namespaces = optional(list(object({
        name = string
        code_handlers = optional(string, "")
        publish_auth_modes = optional(list(string), [])
        subscribe_auth_modes = optional(list(string), [])
        handler_configs = optional(object({
          on_publish = optional(object({
            behavior = optional(string, "")
            data_source_name = string
            lambda_invoke_type = optional(string, "")
          }))
          on_subscribe = optional(object({
            behavior = optional(string, "")
            data_source_name = string
            lambda_invoke_type = optional(string, "")
          }))
        }))
      })), [])
    }))
    datasources = optional(list(object({
      name = string
      description = optional(string, "")
      type = optional(string, "")
      service_role_arn = optional(string, "")
      dynamodb = optional(object({
        table_name = string
        region = optional(string, "")
        use_caller_credentials = optional(bool, false)
        versioned = optional(bool, false)
        delta_sync = optional(object({
          delta_sync_table_name = string
          base_table_ttl = optional(number, 0)
          delta_sync_table_ttl = optional(number, 0)
        }))
      }))
      lambda = optional(object({
        function_arn = string
      }))
      http = optional(object({
        endpoint = string
        sigv4 = optional(object({
          signing_region = optional(string, "")
          signing_service_name = optional(string, "")
        }))
      }))
      opensearch = optional(object({
        endpoint = string
        region = optional(string, "")
      }))
      elasticsearch = optional(object({
        endpoint = string
        region = optional(string, "")
      }))
      eventbridge = optional(object({
        event_bus_arn = string
      }))
      relational_database = optional(object({
        db_cluster_identifier = string
        aws_secret_store_arn = string
        database_name = optional(string, "")
        schema = optional(string, "")
        region = optional(string, "")
      }))
    })), [])
    api_keys = optional(list(object({
      name = string
      description = optional(string, "")
      expires = optional(string, "")
    })), [])
    custom_domain = optional(object({
      domain_name = string
      certificate_arn = string
      description = optional(string, "")
    }))
  })
}