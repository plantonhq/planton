# variables.tf - Terraform variable definitions matching GcpCloudCdnSpec proto

variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name = string
    id = optional(string)
    org  = optional(string)
    env  = optional(string)
    labels = optional(map(string))
    tags = optional(list(string))
    version = optional(object({
      id      = string
      message = string
    }))
  })
}

variable "spec" {
  description = "GCP Cloud CDN specification"
  type = object({
    # Required: GCP Project ID (supports valueFrom references)
    gcp_project_id = object({
      value = string
    })
    
    # Required: Backend configuration
    backend = object({
      # GCS Bucket backend (most common - 80% use case)
      gcs_bucket = optional(object({
        bucket_name           = string
        enable_uniform_access = optional(bool)
      }))
      
      # Compute Engine backend
      compute_service = optional(object({
        instance_group_name = string
        health_check = optional(object({
          path                    = optional(string)
          port                    = optional(number)
          check_interval_seconds  = optional(number)
          timeout_seconds         = optional(number)
          healthy_threshold       = optional(number)
          unhealthy_threshold     = optional(number)
        }))
        protocol = optional(number)  # BackendProtocol enum: 1=HTTP, 2=HTTPS
        port     = optional(number)
      }))
      
      # Cloud Run backend
      cloud_run_service = optional(object({
        service_name = string
        region       = string
      }))
      
      # External/hybrid backend
      external_origin = optional(object({
        hostname = string
        port     = optional(number)
        protocol = optional(number)  # BackendProtocol enum: 1=HTTP, 2=HTTPS
      }))
    })
    
    # Cache mode (enum CacheMode)
    # 0: CACHE_MODE_UNSPECIFIED (will use default: CACHE_ALL_STATIC)
    # 1: CACHE_ALL_STATIC (default, recommended for 90% of deployments)
    # 2: USE_ORIGIN_HEADERS (only cache with explicit Cache-Control headers)
    # 3: FORCE_CACHE_ALL (aggressive caching, use only for public static content)
    cache_mode = optional(number)
    
    # TTL configuration (in seconds)
    default_ttl_seconds = optional(number)  # Default: 3600 (1 hour)
    max_ttl_seconds     = optional(number)  # Default: 86400 (1 day)
    client_ttl_seconds  = optional(number)  # Default: same as max_ttl_seconds
    
    # Negative caching (cache HTTP 4xx/5xx errors)
    enable_negative_caching = optional(bool)  # Default: false
    
    # Advanced configuration (20% use cases)
    advanced_config = optional(object({
      # Cache key policy
      cache_key_policy = optional(object({
        include_query_string   = optional(bool)
        query_string_whitelist = optional(list(string))
        include_protocol       = optional(bool)
        include_host           = optional(bool)
        included_headers       = optional(list(string))
      }))
      
      # Signed URL configuration for private content
      signed_url_config = optional(object({
        enabled = bool
        keys = list(object({
          key_name  = string
          key_value = string
        }))
      }))
      
      # Negative caching policies per status code
      negative_caching_policies = optional(list(object({
        code        = number  # HTTP status code (400-599)
        ttl_seconds = number  # TTL for caching this error
      })))
      
      # Serve-while-stale (seconds)
      serve_while_stale_seconds = optional(number)
      
      # Request coalescing
      enable_request_coalescing = optional(bool)
    }))
    
    # Frontend configuration (SSL, domains, Cloud Armor)
    frontend_config = optional(object({
      # Custom domains
      custom_domains = optional(list(string))
      
      # SSL certificate configuration
      ssl_certificate = optional(object({
        # Google-managed certificate
        google_managed = optional(object({
          domains = list(string)
        }))
        
        # Self-managed certificate
        self_managed = optional(object({
          certificate_pem = string
          private_key_pem = string
        }))
      }))
      
      # Cloud Armor (WAF/DDoS protection)
      cloud_armor = optional(object({
        enabled               = bool
        security_policy_name  = string
      }))
      
      # HTTP to HTTPS redirect
      enable_https_redirect = optional(bool)  # Default: true
    }))
  })
  
  validation {
    condition = var.spec.backend.gcs_bucket != null || var.spec.backend.compute_service != null || var.spec.backend.cloud_run_service != null || var.spec.backend.external_origin != null
    error_message = "At least one backend type must be specified: gcs_bucket, compute_service, cloud_run_service, or external_origin."
  }
  
  validation {
    condition     = length(regexall("^[a-z][a-z0-9-]{4,28}[a-z0-9]$", var.spec.gcp_project_id.value)) > 0
    error_message = "The gcp_project_id.value must be a valid GCP project ID (lowercase, 6-30 characters, start with letter)."
  }
  
  validation {
    condition = var.spec.cache_mode == null || (var.spec.cache_mode >= 0 && var.spec.cache_mode <= 3)
    error_message = "The cache_mode must be a valid CacheMode enum value (0-3)."
  }
  
  validation {
    condition = var.spec.default_ttl_seconds == null || (var.spec.default_ttl_seconds >= 0 && var.spec.default_ttl_seconds <= 31536000)
    error_message = "The default_ttl_seconds must be between 0 and 31536000 (1 year)."
  }
  
  validation {
    condition = var.spec.max_ttl_seconds == null || (var.spec.max_ttl_seconds >= 0 && var.spec.max_ttl_seconds <= 31536000)
    error_message = "The max_ttl_seconds must be between 0 and 31536000 (1 year)."
  }
  
  validation {
    condition = var.spec.client_ttl_seconds == null || (var.spec.client_ttl_seconds >= 0 && var.spec.client_ttl_seconds <= 31536000)
    error_message = "The client_ttl_seconds must be between 0 and 31536000 (1 year)."
  }
}
