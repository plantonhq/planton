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
  description = "DigitalOceanFunction specification"
  type = object({
    function_name = string
    region = string
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
    source_directory = optional(string, "")
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
    project_id = optional(string, "")
    app_name = string
  })
}
