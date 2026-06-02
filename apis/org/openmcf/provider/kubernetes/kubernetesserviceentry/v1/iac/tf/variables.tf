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
  description = "Specification for KubernetesServiceEntry"
  type = object({
    # Namespace the ServiceEntry is created in (resolved foreign key).
    namespace = string

    # Hosts associated with the ServiceEntry. Required, at least one.
    hosts = list(string)

    # Virtual IP addresses or CIDR prefixes the service is reached at. CIDR prefixes are
    # honored only with NONE or STATIC resolution.
    addresses = optional(list(string))

    # Ports exposed by the external service. name and number must each be unique.
    ports = optional(list(object({
      number      = number
      protocol    = optional(string)
      name        = string
      target_port = optional(number)
    })))

    # MESH_EXTERNAL (default) or MESH_INTERNAL.
    location = optional(string)

    # Endpoint resolution mode: NONE (default), STATIC, DNS, or DNS_ROUND_ROBIN.
    resolution = optional(string)

    # Static endpoints backing the service. Mutually exclusive with workload_selector.
    endpoints = optional(list(object({
      address         = optional(string)
      ports           = optional(map(number))
      labels          = optional(map(string))
      network         = optional(string)
      locality        = optional(string)
      weight          = optional(number)
      service_account = optional(string)
    })))

    # Namespaces this service is exported to. Default is all namespaces.
    export_to = optional(list(string))

    # Subject alternate names verified on the server certificate when originating TLS.
    subject_alt_names = optional(list(string))

    # Selects in-mesh workloads by label instead of listing static endpoints. Matched at
    # runtime by istiod; not an OpenMCF foreign key. Mutually exclusive with endpoints.
    workload_selector = optional(object({
      labels = optional(map(string))
    }))
  })
}
