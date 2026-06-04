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
  description = "GCP Cloud Run service specification"
  type = object({
    # Required: GCP project ID where the Cloud Run service will be created
    # Can be a direct value or a reference to a GcpProject resource
    project_id = object({
      value = string
    })

    # Required: GCP region where the service is deployed (e.g., "us-central1")
    region = string

    # Optional: Custom service name (defaults to metadata.name if not provided)
    service_name = optional(string)

    # Optional: Service account email that the Cloud Run service runs as
    service_account = optional(string)

    # Required: Container configuration
    container = object({
      # Container image configuration
      image = object({
        repo = string # Image repository (e.g., "us-docker.pkg.dev/prj/registry/app")
        tag  = string # Image tag (e.g., "1.0.0")
      })

      # Optional: Environment variables and secrets
      env = optional(object({
        variables = optional(map(string)) # Plain environment variables
        secrets   = optional(map(string)) # Secret Manager references
      }))

      # Container port (defaults to 8080)
      port = optional(number, 8080)

      # Required: CPU units (1, 2, or 4)
      cpu = number

      # Required: Memory in MiB (128-32768)
      memory = number

      # Required: Min and max container instances
      replicas = object({
        min = number # Minimum instances (can be 0 for scale-to-zero)
        max = number # Maximum instances
      })
    })

    # Optional: Maximum concurrent requests per instance (1-1000, default 80)
    max_concurrency = optional(number, 80)

    # Optional: Request timeout in seconds (1-3600, default 300)
    timeout_seconds = optional(number, 300)

    # Optional: Ingress settings (default "INGRESS_TRAFFIC_ALL")
    # Valid values: "INGRESS_TRAFFIC_ALL", "INGRESS_TRAFFIC_INTERNAL_ONLY", "INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER"
    ingress = optional(string, "INGRESS_TRAFFIC_ALL")

    # Optional: Allow unauthenticated access (default true)
    allow_unauthenticated = optional(bool, true)

    # Optional: VPC access configuration for private resource access
    vpc_access = optional(object({
      # VPC network name - can be a direct value or a reference to a GcpVpc resource
      network = optional(object({
        value = string
      }))
      # VPC subnet name - can be a direct value or a reference to a GcpSubnetwork resource
      subnet = optional(object({
        value = string
      }))
      egress = optional(string) # "ALL_TRAFFIC" or "PRIVATE_RANGES_ONLY"
    }))

    # Optional: Execution environment (default "EXECUTION_ENVIRONMENT_GEN2")
    # Valid values: "EXECUTION_ENVIRONMENT_GEN1", "EXECUTION_ENVIRONMENT_GEN2"
    execution_environment = optional(string, "EXECUTION_ENVIRONMENT_GEN2")

    # Optional: Custom DNS mapping
    dns = optional(object({
      enabled   = bool         # Enable custom domain mapping
      hostnames = list(string) # List of custom hostnames
      # Cloud DNS managed zone for verification. StringValueOrRef: the platform
      # resolves any reference to a literal before apply, so only `value` arrives.
      managed_zone = object({
        value = string
      })
    }))

    # Optional: Deletion protection to prevent accidental deletion (default false)
    delete_protection = optional(bool, false)
  })
}
