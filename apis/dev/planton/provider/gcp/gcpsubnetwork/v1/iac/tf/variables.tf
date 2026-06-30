variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for the GCP Subnetwork"
  type = object({
    # The GCP project ID in which to create this subnetwork
    project_id = string

    # Reference to the parent VPC network (self-link)
    # Example: "projects/my-project/global/networks/my-vpc"
    vpc_self_link = string

    # Region in which to create this subnet
    # Example: "us-central1"
    region = string

    # Primary IPv4 CIDR range for the subnet
    # Example: "10.0.0.0/20"
    ip_cidr_range = string

    # Secondary IP ranges for alias IPs (e.g., for GKE Pod or Service IPs)
    secondary_ip_ranges = optional(list(object({
      range_name    = string
      ip_cidr_range = string
    })))

    # Whether to enable Private Google Access on this subnet
    private_ip_google_access = optional(bool)

    # Name of the subnetwork to create in GCP
    subnetwork_name = string
  })

  validation {
    condition     = can(regex("^[a-z]([-a-z0-9]*[a-z0-9])?$", var.spec.region))
    error_message = "Region must be a valid GCP region name (lowercase letters, numbers, and hyphens)."
  }

  validation {
    condition     = can(regex("^\\d+\\.\\d+\\.\\d+\\.\\d+/\\d+$", var.spec.ip_cidr_range))
    error_message = "ip_cidr_range must be a valid IPv4 CIDR notation (e.g., 10.0.0.0/20)."
  }

  validation {
    condition     = can(regex("^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$", var.spec.subnetwork_name))
    error_message = "Subnetwork name must be 1-63 characters, lowercase letters, numbers, or hyphens, starting with a letter and ending with a letter or number."
  }
}
