variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Scaleway Load Balancer specification"
  type = object({
    # Zone where the LB will be created (e.g., "fr-par-1")
    zone = string

    # LB type: "LB-S", "LB-GP-M", "LB-GP-L", "LB-GP-XL"
    type = string

    # Private Network ID (resolved from StringValueOrRef before Terraform runs)
    private_network_id = optional(string, "")

    # Human-readable description
    description = optional(string, "")

    # Minimum SSL/TLS compatibility level
    ssl_compatibility_level = optional(string, "")

    # Backend server pools
    backends = list(object({
      name                       = string
      server_ips                 = list(string)
      forward_port               = number
      forward_protocol           = string
      forward_port_algorithm     = optional(string, "roundrobin")
      sticky_sessions            = optional(string, "none")
      sticky_sessions_cookie_name = optional(string, "")
      timeout_connect            = optional(string, "")
      timeout_server             = optional(string, "")
      on_marked_down_action      = optional(string, "")
      ssl_bridging               = optional(bool, false)
      proxy_protocol             = optional(string, "none")
      health_check = optional(object({
        type              = optional(string, "tcp")
        uri               = optional(string, "/")
        expected_code     = optional(number, 200)
        check_delay       = optional(string, "5s")
        check_timeout     = optional(string, "3s")
        check_max_retries = optional(number, 3)
        port              = optional(number, 0)
      }))
    }))

    # Frontend listeners
    frontends = list(object({
      name              = string
      inbound_port      = number
      backend_name      = string
      certificate_names = optional(list(string), [])
      timeout_client    = optional(string, "")
      enable_http3      = optional(bool, false)
    }))

    # TLS certificates
    certificates = optional(list(object({
      name = string
      letsencrypt = optional(object({
        common_name              = string
        subject_alternative_names = optional(list(string), [])
      }))
      custom_certificate = optional(object({
        certificate_chain = string
      }))
    })), [])
  })
}

variable "scaleway_access_key" {
  description = "Scaleway access key for API authentication"
  type        = string
  sensitive   = true
}

variable "scaleway_secret_key" {
  description = "Scaleway secret key for API authentication"
  type        = string
  sensitive   = true
}

variable "scaleway_project_id" {
  description = "Scaleway project ID (optional, defaults from provider)"
  type        = string
  default     = ""
}

variable "scaleway_organization_id" {
  description = "Scaleway organization ID (optional, defaults from provider)"
  type        = string
  default     = ""
}
