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
  description = "HetznerCloudLoadBalancer specification"
  type = object({
    load_balancer_type = string
    location           = string
    algorithm          = optional(string)
    delete_protection  = optional(bool)

    services = list(object({
      protocol         = string
      listen_port      = optional(number)
      destination_port = optional(number)
      proxyprotocol    = optional(bool)

      http = optional(object({
        sticky_sessions = optional(bool)
        cookie_name     = optional(string)
        cookie_lifetime = optional(number)
        certificate_ids = optional(list(string))
        redirect_http   = optional(bool)
      }))

      health_check = optional(object({
        protocol = optional(string)
        port     = optional(number)
        interval = optional(number)
        timeout  = optional(number)
        retries  = optional(number)

        http = optional(object({
          domain       = optional(string)
          path         = optional(string)
          response     = optional(string)
          tls          = optional(bool)
          status_codes = optional(list(string))
        }))
      }))
    }))

    server_targets = optional(list(object({
      server_id      = string
      use_private_ip = optional(bool)
    })))

    label_selector_targets = optional(list(object({
      selector       = string
      use_private_ip = optional(bool)
    })))

    ip_targets = optional(list(object({
      ip = string
    })))

    network = optional(object({
      network_id              = string
      ip                      = optional(string)
      enable_public_interface = optional(bool)
    }))
  })
}

variable "hcloud_token" {
  description = "Hetzner Cloud API token for authentication"
  type        = string
  sensitive   = true
}
