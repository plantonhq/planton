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
  description = "Specification for the GCP Cloud Router NAT"
  type = object({
    project_id             = string
    vpc_self_link          = string
    region                 = string
    subnetwork_self_links  = optional(list(string), [])
    nat_ip_names           = optional(list(string), [])
    log_filter             = optional(string, "ERRORS_ONLY") # DISABLED, ERRORS_ONLY, or ALL
    router_name            = string
    nat_name               = string
  })
  
  validation {
    condition     = can(regex("^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$", var.spec.router_name))
    error_message = "Router name must be 1-63 characters, lowercase letters, numbers, or hyphens, starting with a letter and ending with a letter or number."
  }
  
  validation {
    condition     = can(regex("^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$", var.spec.nat_name))
    error_message = "NAT name must be 1-63 characters, lowercase letters, numbers, or hyphens, starting with a letter and ending with a letter or number."
  }
}
