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
  description = "AwsRdsProxy specification"
  type = object({
    region = string
    engine_family = optional(string, "")
    role_arn = string
    vpc_subnet_ids = list(string)
    vpc_security_group_ids = optional(list(string), [])
    auth = list(object({
      secret_arn = string
      description = optional(string, "")
      iam_auth = optional(string, "")
      client_password_auth_type = optional(string, "")
      username = optional(string, "")
    }))
    require_tls = optional(bool, false)
    idle_client_timeout = optional(number, 0)
    debug_logging = optional(bool, false)
    default_auth_scheme = optional(string, "")
    endpoint_network_type = optional(string, "")
    target_connection_network_type = optional(string, "")
    connection_pool = optional(object({
      connection_borrow_timeout = optional(number)
      init_query = optional(string, "")
      max_connections_percent = optional(number)
      max_idle_connections_percent = optional(number)
      session_pinning_filters = optional(list(string), [])
    }))
    endpoints = optional(list(object({
      name = string
      target_role = optional(string, "")
      vpc_subnet_ids = list(string)
      vpc_security_group_ids = optional(list(string), [])
    })), [])
    target = optional(object({
      db_instance_identifier = optional(string, "")
      db_cluster_identifier = optional(string, "")
    }))
  })
}