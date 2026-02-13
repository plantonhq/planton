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
  description = "Scaleway Instance specification"
  type = object({
    # Zone where the instance will be created (e.g., "fr-par-1")
    zone = string

    # Instance commercial type (e.g., "DEV1-S", "PRO2-M")
    type = string

    # Base image UUID or label (e.g., "ubuntu_focal")
    image = string

    # Public IP configuration (null = no public IP)
    public_ip = optional(object({
      reverse_dns = optional(string, "")
    }))

    # Security group ID (resolved from StringValueOrRef before Terraform runs)
    security_group_id = optional(string, "")

    # Private Network ID (resolved from StringValueOrRef before Terraform runs)
    private_network_id = optional(string, "")

    # Root volume configuration
    root_volume = optional(object({
      size_in_gb             = optional(number)
      volume_type            = optional(string)
      delete_on_termination  = optional(bool, true)
      sbs_iops               = optional(number)
    }))

    # Additional local volumes to create and attach
    additional_volumes = optional(list(object({
      name        = optional(string, "")
      volume_type = string
      size_in_gb  = number
    })), [])

    # Cloud-init script
    cloud_init = optional(string, "")

    # Instance state: "started", "stopped", "standby"
    state = optional(string, "started")

    # Deletion protection
    protected = optional(bool, false)
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
