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
  description = "DigitalOcean App Platform Service specification"
  type = object({
    service_name  = string
    region        = string
    service_type  = string # "web_service", "worker", or "job"
    
    # Source configuration - must provide exactly one
    git_source = optional(object({
      repo_url      = string
      branch        = string
      build_command = optional(string)
      run_command   = optional(string)
    }))
    
    image_source = optional(object({
      registry   = string
      repository = string
      tag        = string
    }))
    
    instance_size_slug = string
    instance_count     = optional(number, 1)
    
    enable_autoscale   = optional(bool, false)
    min_instance_count = optional(number)
    max_instance_count = optional(number)
    
    env = optional(map(string), {})
    
    custom_domain = optional(string)
  })
  
  validation {
    condition     = contains(["web_service", "worker", "job"], var.spec.service_type)
    error_message = "service_type must be one of: web_service, worker, job"
  }
  
  validation {
    condition     = (var.spec.git_source != null && var.spec.image_source == null) || (var.spec.git_source == null && var.spec.image_source != null)
    error_message = "Must specify exactly one of git_source or image_source"
  }
  
  validation {
    condition     = !var.spec.enable_autoscale || (var.spec.min_instance_count != null && var.spec.max_instance_count != null)
    error_message = "When enable_autoscale is true, both min_instance_count and max_instance_count must be specified"
  }
}

variable "digitalocean_token" {
  description = "DigitalOcean API token for authentication"
  type        = string
  sensitive   = true
}
