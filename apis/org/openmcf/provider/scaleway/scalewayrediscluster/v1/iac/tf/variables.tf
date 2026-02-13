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
  description = "Scaleway Redis Cluster specification"
  type = object({
    # Zone where the cluster will be created (e.g., "fr-par-1")
    zone = string

    # Redis version (e.g., "7.2.5", "6.2.7")
    version = string

    # Node type (e.g., "RED1-MICRO", "RED1-M")
    node_type = string

    # Cluster size: 1=standalone, 2=HA, 3+=cluster
    cluster_size = optional(number, 1)

    # TLS encryption (forces recreation if changed)
    tls_enabled = optional(bool, false)

    # Authentication
    user_name = string
    password  = string

    # ACL rules (mutually exclusive with private_network_id)
    acl_rules = optional(list(object({
      ip          = string
      description = optional(string, "")
    })), [])

    # Private Network ID (resolved from StringValueOrRef before Terraform runs)
    # Mutually exclusive with acl_rules
    private_network_id = optional(string, "")

    # Redis settings
    settings = optional(map(string), {})
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
