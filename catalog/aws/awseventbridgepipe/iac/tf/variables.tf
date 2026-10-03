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
  description = "AwsEventBridgePipe specification"
  type = object({
    region = string
    description = optional(string, "")
    source = string
    source_parameters = optional(object({
      filter_criteria = optional(object({
        filters = list(object({
          pattern = string
        }))
      }))
      sqs = optional(object({
        batch_size = optional(number)
        maximum_batching_window_in_seconds = optional(number)
      }))
      kinesis = optional(object({
        starting_position = optional(string, "")
        starting_position_timestamp = optional(string, "")
        batch_size = optional(number)
        maximum_batching_window_in_seconds = optional(number)
        maximum_record_age_in_seconds = optional(number)
        maximum_retry_attempts = optional(number)
        on_partial_batch_item_failure = optional(string, "")
        parallelization_factor = optional(number)
        dead_letter_queue_arn = optional(string, "")
      }))
      dynamodb = optional(object({
        starting_position = optional(string, "")
        batch_size = optional(number)
        maximum_batching_window_in_seconds = optional(number)
        maximum_record_age_in_seconds = optional(number)
        maximum_retry_attempts = optional(number)
        on_partial_batch_item_failure = optional(string, "")
        parallelization_factor = optional(number)
        dead_letter_queue_arn = optional(string, "")
      }))
      msk = optional(object({
        topic_name = string
        consumer_group_id = optional(string, "")
        starting_position = optional(string, "")
        batch_size = optional(number)
        maximum_batching_window_in_seconds = optional(number)
        credentials = optional(object({
          client_certificate_tls_auth = optional(string, "")
          sasl_scram_512_auth = optional(string, "")
        }))
      }))
      self_managed_kafka = optional(object({
        topic_name = string
        additional_bootstrap_servers = optional(list(string), [])
        consumer_group_id = optional(string, "")
        starting_position = optional(string, "")
        batch_size = optional(number)
        maximum_batching_window_in_seconds = optional(number)
        credentials = optional(object({
          basic_auth = optional(string, "")
          client_certificate_tls_auth = optional(string, "")
          sasl_scram_256_auth = optional(string, "")
          sasl_scram_512_auth = optional(string, "")
        }))
        server_root_ca_certificate = optional(string, "")
        vpc = optional(object({
          subnets = list(string)
          security_groups = optional(list(string), [])
        }))
      }))
      activemq = optional(object({
        queue_name = string
        basic_auth_credentials = optional(string, "")
        batch_size = optional(number)
        maximum_batching_window_in_seconds = optional(number)
      }))
      rabbitmq = optional(object({
        queue_name = string
        virtual_host = optional(string, "")
        basic_auth_credentials = optional(string, "")
        batch_size = optional(number)
        maximum_batching_window_in_seconds = optional(number)
      }))
    }))
    enrichment = optional(string, "")
    enrichment_parameters = optional(object({
      input_template = optional(string, "")
      http_parameters = optional(object({
        header_parameters = optional(map(string), {})
        path_parameter_value = optional(string, "")
        query_string_parameters = optional(map(string), {})
      }))
    }))
    target = string
    target_parameters = optional(object({
      input_template = optional(string, "")
      sqs = optional(object({
        message_group_id = optional(string, "")
        message_deduplication_id = optional(string, "")
      }))
      kinesis = optional(object({
        partition_key = string
      }))
      lambda = optional(object({
        invocation_type = optional(string, "")
      }))
      step_function = optional(object({
        invocation_type = optional(string, "")
      }))
      ecs_task = optional(object({
        task_definition_arn = string
        task_count = optional(number)
        launch_type = optional(string, "")
        capacity_provider_strategy = optional(list(object({
          capacity_provider = string
          base = optional(number, 0)
          weight = optional(number, 0)
        })), [])
        network_configuration = optional(object({
          subnets = list(string)
          security_groups = optional(list(string), [])
          assign_public_ip = optional(bool, false)
        }))
        group = optional(string, "")
        platform_version = optional(string, "")
        overrides = optional(object({
          container_overrides = optional(list(object({
            name = optional(string, "")
            command = optional(list(string), [])
            cpu = optional(number)
            memory = optional(number)
            memory_reservation = optional(number)
            environment = optional(list(object({
              name = string
              value = optional(string, "")
            })), [])
            environment_files = optional(list(object({
              type = optional(string, "")
              value = optional(string, "")
            })), [])
            resource_requirements = optional(list(object({
              type = optional(string, "")
              value = string
            })), [])
          })), [])
          cpu = optional(string, "")
          memory = optional(string, "")
          ephemeral_storage_size_in_gib = optional(number)
          execution_role_arn = optional(string, "")
          task_role_arn = optional(string, "")
          inference_accelerator_overrides = optional(list(object({
            device_name = optional(string, "")
            device_type = optional(string, "")
          })), [])
        }))
        placement_constraints = optional(list(object({
          type = optional(string, "")
          expression = optional(string, "")
        })), [])
        placement_strategy = optional(list(object({
          type = optional(string, "")
          field = optional(string, "")
        })), [])
        propagate_tags = optional(string, "")
        tags = optional(map(string), {})
        enable_ecs_managed_tags = optional(bool, false)
        enable_execute_command = optional(bool, false)
        reference_id = optional(string, "")
      }))
      batch_job = optional(object({
        job_definition = string
        job_name = string
        array_size = optional(number)
        retry_attempts = optional(number)
        parameters = optional(map(string), {})
        depends_on = optional(list(object({
          job_id = optional(string, "")
          type = optional(string, "")
        })), [])
        container_overrides = optional(object({
          command = optional(list(string), [])
          environment = optional(list(object({
            name = string
            value = optional(string, "")
          })), [])
          instance_type = optional(string, "")
          resource_requirements = optional(list(object({
            type = optional(string, "")
            value = string
          })), [])
        }))
      }))
      redshift_data = optional(object({
        database = string
        sqls = list(string)
        db_user = optional(string, "")
        secret_manager_arn = optional(string, "")
        statement_name = optional(string, "")
        with_event = optional(bool, false)
      }))
      sagemaker_pipeline = optional(object({
        pipeline_parameters = optional(list(object({
          name = string
          value = string
        })), [])
      }))
      eventbridge_event_bus = optional(object({
        detail_type = optional(string, "")
        source = optional(string, "")
        endpoint_id = optional(string, "")
        resources = optional(list(string), [])
        time = optional(string, "")
      }))
      cloudwatch_logs = optional(object({
        log_stream_name = optional(string, "")
        timestamp = optional(string, "")
      }))
      http = optional(object({
        header_parameters = optional(map(string), {})
        path_parameter_value = optional(string, "")
        query_string_parameters = optional(map(string), {})
      }))
    }))
    role_arn = string
    desired_state = optional(string, "")
    kms_key_identifier = optional(string, "")
    log_configuration = optional(object({
      level = optional(string, "")
      include_execution_data = optional(bool, false)
      cloudwatch_logs = optional(object({
        log_group_arn = string
      }))
      firehose = optional(object({
        delivery_stream_arn = string
      }))
      s3 = optional(object({
        bucket_name = string
        bucket_owner = optional(string, "")
        output_format = optional(string, "")
        prefix = optional(string, "")
      }))
    }))
  })
}