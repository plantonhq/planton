# Input variables for Kubernetes Namespace Terraform module

variable "metadata" {
  description = "Metadata for the namespace resource"
  type = object({
    name = string
    org  = optional(string)
    env  = optional(string)
  })
}

variable "spec" {
  description = "Specification for the Kubernetes namespace"
  type = object({
    name = string
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    
    resource_profile = optional(object({
      preset = optional(string)
      custom = optional(object({
        cpu = optional(object({
          requests = string
          limits   = string
        }))
        memory = optional(object({
          requests = string
          limits   = string
        }))
        object_counts = optional(object({
          pods                      = optional(number)
          services                  = optional(number)
          configmaps                = optional(number)
          secrets                   = optional(number)
          persistent_volume_claims  = optional(number)
          load_balancers            = optional(number)
        }))
        default_limits = optional(object({
          default_cpu_request    = string
          default_cpu_limit      = string
          default_memory_request = string
          default_memory_limit   = string
        }))
      }))
    }))

    network_config = optional(object({
      isolate_ingress              = optional(bool, false)
      restrict_egress              = optional(bool, false)
      allowed_ingress_namespaces   = optional(list(string), [])
      allowed_egress_cidrs         = optional(list(string), [])
      allowed_egress_domains       = optional(list(string), [])
    }))

    service_mesh_config = optional(object({
      enabled      = optional(bool, false)
      mesh_type    = optional(string)
      revision_tag = optional(string)
    }))

    pod_security_standard = optional(string)
  })
}


