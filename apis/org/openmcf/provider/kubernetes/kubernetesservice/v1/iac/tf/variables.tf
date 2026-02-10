# Input variables for KubernetesService Terraform module
# These mirror the KubernetesServiceSpec protobuf schema.

variable "metadata" {
  description = "Metadata for the service resource"
  type = object({
    name = string
    org  = optional(string)
    env  = optional(string)
  })
}

variable "spec" {
  description = "Specification for the Kubernetes Service"
  type = object({
    namespace = string
    name      = string
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})

    type     = optional(string, "cluster_ip")
    selector = optional(map(string), {})

    ports = optional(list(object({
      name        = optional(string, "")
      protocol    = optional(string, "TCP")
      port        = number
      target_port = optional(string, "")
      node_port   = optional(number, 0)
    })), [])

    headless         = optional(bool, false)
    external_dns_name = optional(string, "")

    external_traffic_policy = optional(string, "cluster")
    session_affinity        = optional(string, "none")

    load_balancer_source_ranges = optional(list(string), [])
  })
}
