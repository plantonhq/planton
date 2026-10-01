variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "DigitalOceanApp specification"
  type = object({
    app_name = string
    region = string
    services = optional(list(object({
      name = string
      source_dir = optional(string, "")
      git = optional(object({
        repo_clone_url = string
        branch = string
      }))
      github = optional(object({
        repo = string
        branch = string
        deploy_on_push = optional(bool, false)
      }))
      gitlab = optional(object({
        repo = string
        branch = string
        deploy_on_push = optional(bool, false)
      }))
      bitbucket = optional(object({
        repo = string
        branch = string
        deploy_on_push = optional(bool, false)
      }))
      image = optional(object({
        registry_type = string
        registry = optional(string, "")
        repository = string
        tag = optional(string, "")
        digest = optional(string, "")
        registry_credentials = optional(string, "")
        deploy_on_push = optional(bool, false)
      }))
      environment_slug = optional(string, "")
      dockerfile_path = optional(string, "")
      build_command = optional(string, "")
      run_command = optional(string, "")
      instance_size_slug = optional(string, "")
      instance_count = optional(number, 0)
      http_port = optional(number)
      internal_ports = optional(list(number), [])
      health_check = optional(object({
        port = optional(number)
        http_path = optional(string, "")
        initial_delay_seconds = optional(number)
        period_seconds = optional(number)
        timeout_seconds = optional(number)
        success_threshold = optional(number)
        failure_threshold = optional(number)
      }))
      liveness_health_check = optional(object({
        port = optional(number)
        http_path = optional(string, "")
        initial_delay_seconds = optional(number)
        period_seconds = optional(number)
        timeout_seconds = optional(number)
        success_threshold = optional(number)
        failure_threshold = optional(number)
      }))
      autoscaling = optional(object({
        min_instance_count = number
        max_instance_count = number
        cpu_percent = number
      }))
      termination = optional(object({
        grace_period_seconds = optional(number)
        drain_seconds = optional(number)
      }))
      envs = optional(list(object({
        key = string
        plaintext = optional(string, "")
        secret = optional(string, "")
        scope = optional(string, "")
      })), [])
      alerts = optional(list(object({
        rule = string
        operator = string
        window = string
        value = optional(number, 0)
        disabled = optional(bool, false)
        destinations = optional(object({
          emails = optional(list(string), [])
          slack_webhooks = optional(list(object({
            channel = string
            url = string
          })), [])
        }))
      })), [])
      log_destinations = optional(list(object({
        name = string
        papertrail = optional(object({
          endpoint = string
        }))
        datadog = optional(object({
          api_key = string
          endpoint = optional(string, "")
        }))
        logtail = optional(object({
          token = string
        }))
        open_search = optional(object({
          endpoint = optional(string, "")
          index_name = optional(string, "")
          cluster_name = optional(string, "")
          basic_auth = optional(object({
            user = optional(string, "")
            password = optional(string, "")
          }))
        }))
      })), [])
    })), [])
    workers = optional(list(object({
      name = string
      source_dir = optional(string, "")
      git = optional(object({
        repo_clone_url = string
        branch = string
      }))
      github = optional(object({
        repo = string
        branch = string
        deploy_on_push = optional(bool, false)
      }))
      gitlab = optional(object({
        repo = string
        branch = string
        deploy_on_push = optional(bool, false)
      }))
      bitbucket = optional(object({
        repo = string
        branch = string
        deploy_on_push = optional(bool, false)
      }))
      image = optional(object({
        registry_type = string
        registry = optional(string, "")
        repository = string
        tag = optional(string, "")
        digest = optional(string, "")
        registry_credentials = optional(string, "")
        deploy_on_push = optional(bool, false)
      }))
      environment_slug = optional(string, "")
      dockerfile_path = optional(string, "")
      build_command = optional(string, "")
      run_command = optional(string, "")
      instance_size_slug = optional(string, "")
      instance_count = optional(number, 0)
      liveness_health_check = optional(object({
        port = optional(number)
        http_path = optional(string, "")
        initial_delay_seconds = optional(number)
        period_seconds = optional(number)
        timeout_seconds = optional(number)
        success_threshold = optional(number)
        failure_threshold = optional(number)
      }))
      autoscaling = optional(object({
        min_instance_count = number
        max_instance_count = number
        cpu_percent = number
      }))
      termination = optional(object({
        grace_period_seconds = optional(number)
        drain_seconds = optional(number)
      }))
      envs = optional(list(object({
        key = string
        plaintext = optional(string, "")
        secret = optional(string, "")
        scope = optional(string, "")
      })), [])
      alerts = optional(list(object({
        rule = string
        operator = string
        window = string
        value = optional(number, 0)
        disabled = optional(bool, false)
        destinations = optional(object({
          emails = optional(list(string), [])
          slack_webhooks = optional(list(object({
            channel = string
            url = string
          })), [])
        }))
      })), [])
      log_destinations = optional(list(object({
        name = string
        papertrail = optional(object({
          endpoint = string
        }))
        datadog = optional(object({
          api_key = string
          endpoint = optional(string, "")
        }))
        logtail = optional(object({
          token = string
        }))
        open_search = optional(object({
          endpoint = optional(string, "")
          index_name = optional(string, "")
          cluster_name = optional(string, "")
          basic_auth = optional(object({
            user = optional(string, "")
            password = optional(string, "")
          }))
        }))
      })), [])
    })), [])
    jobs = optional(list(object({
      name = string
      source_dir = optional(string, "")
      git = optional(object({
        repo_clone_url = string
        branch = string
      }))
      github = optional(object({
        repo = string
        branch = string
        deploy_on_push = optional(bool, false)
      }))
      gitlab = optional(object({
        repo = string
        branch = string
        deploy_on_push = optional(bool, false)
      }))
      bitbucket = optional(object({
        repo = string
        branch = string
        deploy_on_push = optional(bool, false)
      }))
      image = optional(object({
        registry_type = string
        registry = optional(string, "")
        repository = string
        tag = optional(string, "")
        digest = optional(string, "")
        registry_credentials = optional(string, "")
        deploy_on_push = optional(bool, false)
      }))
      environment_slug = optional(string, "")
      dockerfile_path = optional(string, "")
      build_command = optional(string, "")
      run_command = optional(string, "")
      instance_size_slug = optional(string, "")
      instance_count = optional(number, 0)
      kind = optional(string, "")
      termination = optional(object({
        grace_period_seconds = optional(number)
        drain_seconds = optional(number)
      }))
      envs = optional(list(object({
        key = string
        plaintext = optional(string, "")
        secret = optional(string, "")
        scope = optional(string, "")
      })), [])
      alerts = optional(list(object({
        rule = string
        operator = string
        window = string
        value = optional(number, 0)
        disabled = optional(bool, false)
        destinations = optional(object({
          emails = optional(list(string), [])
          slack_webhooks = optional(list(object({
            channel = string
            url = string
          })), [])
        }))
      })), [])
      log_destinations = optional(list(object({
        name = string
        papertrail = optional(object({
          endpoint = string
        }))
        datadog = optional(object({
          api_key = string
          endpoint = optional(string, "")
        }))
        logtail = optional(object({
          token = string
        }))
        open_search = optional(object({
          endpoint = optional(string, "")
          index_name = optional(string, "")
          cluster_name = optional(string, "")
          basic_auth = optional(object({
            user = optional(string, "")
            password = optional(string, "")
          }))
        }))
      })), [])
    })), [])
    static_sites = optional(list(object({
      name = string
      source_dir = optional(string, "")
      git = optional(object({
        repo_clone_url = string
        branch = string
      }))
      github = optional(object({
        repo = string
        branch = string
        deploy_on_push = optional(bool, false)
      }))
      gitlab = optional(object({
        repo = string
        branch = string
        deploy_on_push = optional(bool, false)
      }))
      bitbucket = optional(object({
        repo = string
        branch = string
        deploy_on_push = optional(bool, false)
      }))
      environment_slug = optional(string, "")
      dockerfile_path = optional(string, "")
      build_command = optional(string, "")
      output_dir = optional(string, "")
      index_document = optional(string, "")
      error_document = optional(string, "")
      catchall_document = optional(string, "")
      envs = optional(list(object({
        key = string
        plaintext = optional(string, "")
        secret = optional(string, "")
        scope = optional(string, "")
      })), [])
    })), [])
    functions = optional(list(object({
      name = string
      source_dir = optional(string, "")
      git = optional(object({
        repo_clone_url = string
        branch = string
      }))
      github = optional(object({
        repo = string
        branch = string
        deploy_on_push = optional(bool, false)
      }))
      gitlab = optional(object({
        repo = string
        branch = string
        deploy_on_push = optional(bool, false)
      }))
      bitbucket = optional(object({
        repo = string
        branch = string
        deploy_on_push = optional(bool, false)
      }))
      envs = optional(list(object({
        key = string
        plaintext = optional(string, "")
        secret = optional(string, "")
        scope = optional(string, "")
      })), [])
      alerts = optional(list(object({
        rule = string
        operator = string
        window = string
        value = optional(number, 0)
        disabled = optional(bool, false)
        destinations = optional(object({
          emails = optional(list(string), [])
          slack_webhooks = optional(list(object({
            channel = string
            url = string
          })), [])
        }))
      })), [])
      log_destinations = optional(list(object({
        name = string
        papertrail = optional(object({
          endpoint = string
        }))
        datadog = optional(object({
          api_key = string
          endpoint = optional(string, "")
        }))
        logtail = optional(object({
          token = string
        }))
        open_search = optional(object({
          endpoint = optional(string, "")
          index_name = optional(string, "")
          cluster_name = optional(string, "")
          basic_auth = optional(object({
            user = optional(string, "")
            password = optional(string, "")
          }))
        }))
      })), [])
    })), [])
    databases = optional(list(object({
      name = optional(string, "")
      engine = optional(string, "")
      version = optional(string, "")
      production = optional(bool, false)
      cluster_name = optional(string, "")
      db_name = optional(string, "")
      db_user = optional(string, "")
    })), [])
    domains = optional(list(object({
      name = string
      wildcard = optional(bool, false)
      zone = optional(string, "")
      type = optional(string, "")
    })), [])
    envs = optional(list(object({
      key = string
      plaintext = optional(string, "")
      secret = optional(string, "")
      scope = optional(string, "")
    })), [])
    alerts = optional(list(object({
      rule = string
      disabled = optional(bool, false)
      destinations = optional(object({
        emails = optional(list(string), [])
        slack_webhooks = optional(list(object({
          channel = string
          url = string
        })), [])
      }))
    })), [])
    ingress = optional(object({
      rules = optional(list(object({
        match = optional(object({
          path_prefix = optional(string, "")
          authority_exact = optional(string, "")
        }))
        component = optional(object({
          name = string
          preserve_path_prefix = optional(bool, false)
          rewrite = optional(string, "")
        }))
        redirect = optional(object({
          uri = optional(string, "")
          authority = optional(string, "")
          port = optional(number)
          scheme = optional(string, "")
          redirect_code = optional(number)
        }))
        cors = optional(object({
          allow_origins = optional(object({
            exact = optional(string, "")
            regex = optional(string, "")
          }))
          allow_methods = optional(list(string), [])
          allow_headers = optional(list(string), [])
          expose_headers = optional(list(string), [])
          max_age = optional(string, "")
          allow_credentials = optional(bool, false)
        }))
      })), [])
      secure_header = optional(object({
        key = string
        value = string
      }))
    }))
    egress = optional(string, "")
    maintenance = optional(object({
      enabled = optional(bool, false)
      archive = optional(bool, false)
      offline_page_url = optional(string, "")
    }))
    vpc = optional(string, "")
    features = optional(list(string), [])
    disable_edge_cache = optional(bool, false)
    disable_email_obfuscation = optional(bool, false)
    enhanced_threat_control_enabled = optional(bool, false)
    project_id = optional(string, "")
  })
}
