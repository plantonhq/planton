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
  description = "AwsEcsService specification"
  type = object({
    region = string
    cluster_arn = string
    container = object({
      image = optional(object({
        repo = optional(string, "")
        tag = optional(string, "")
      }))
      env = optional(object({
        variables = optional(map(string), {})
        secrets = optional(map(string), {})
        s3_files = optional(list(string), [])
      }))
      port = optional(number, 0)
      replicas = optional(number, 0)
      cpu = number
      memory = number
      logging = optional(object({
        enabled = optional(bool, false)
      }))
    })
    network = object({
      subnets = list(string)
      security_groups = optional(list(string), [])
    })
    iam = optional(object({
      task_execution_role_arn = optional(string, "")
      task_role_arn = optional(string, "")
    }))
    alb = optional(object({
      enabled = optional(bool, false)
      arn = optional(string, "")
      routing_type = optional(string, "")
      path = optional(string, "")
      hostname = optional(string, "")
      listener_port = number
      listener_priority = optional(number, 0)
      health_check = optional(object({
        protocol = optional(string, "")
        path = optional(string, "")
        port = optional(string, "")
        interval = optional(number, 0)
        timeout = optional(number, 0)
        healthy_threshold = optional(number, 0)
        unhealthy_threshold = optional(number, 0)
      }))
    }))
    health_check_grace_period_seconds = optional(number, 0)
    autoscaling = optional(object({
      enabled = optional(bool, false)
      min_tasks = optional(number, 0)
      max_tasks = optional(number, 0)
      target_cpu_percent = optional(number, 0)
      target_memory_percent = optional(number, 0)
    }))
  })
}
