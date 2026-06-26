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
  description = "Specification for Cloudflare R2 Bucket"
  # NOTE: every scalar is optional because the proto->tfvars converter
  # (protojson, EmitUnpopulated:false) omits default/empty scalar values.
  type = object({
    # DNS-compatible bucket name (3-63 characters, lowercase alphanumeric + hyphens)
    bucket_name = string

    # Cloudflare account ID (32 hex characters)
    account_id = string

    # Primary region for the bucket (location hint). One of "wnam", "enam",
    # "weur", "eeur", "apac", "oc". Omitted means "auto" (Cloudflare chooses).
    location = optional(string)

    # Data-residency jurisdiction: "default", "eu", or "fedramp". Omitted = "default".
    jurisdiction = optional(string)

    # Default storage class for new objects: "Standard" or "InfrequentAccess".
    storage_class = optional(string)

    # Expose the bucket via the managed r2.dev public URL.
    public_access = optional(bool, false)

    # Custom domains serving the bucket over your own hostnames.
    custom_domains = optional(list(object({
      enabled = optional(bool, false)
      zone_id = optional(string)
      domain  = optional(string, "")
      min_tls = optional(string)
      ciphers = optional(list(string))
    })), [])

    # CORS configuration.
    cors = optional(object({
      rules = optional(list(object({
        allowed = object({
          methods = list(string)
          origins = list(string)
          headers = optional(list(string))
        })
        id              = optional(string)
        expose_headers  = optional(list(string))
        max_age_seconds = optional(number)
      })), [])
    }))

    # Object-lifecycle configuration.
    lifecycle = optional(object({
      rules = optional(list(object({
        id         = string
        conditions = object({ prefix = optional(string, "") })
        enabled    = optional(bool, false)
        abort_multipart_uploads_transition = optional(object({
          max_age_seconds = number
        }))
        delete_objects_transition = optional(object({
          condition = object({
            type            = string
            max_age_seconds = optional(number)
            date            = optional(string)
          })
        }))
        storage_class_transitions = optional(list(object({
          condition = object({
            type            = string
            max_age_seconds = optional(number)
            date            = optional(string)
          })
        })), [])
      })), [])
    }))

    # Object-lock (retention) configuration.
    lock = optional(object({
      rules = optional(list(object({
        id = string
        condition = object({
          type            = string
          max_age_seconds = optional(number)
          date            = optional(string)
        })
        enabled = optional(bool, false)
        prefix  = optional(string, "")
      })), [])
    }))

    # Event notifications: forward object events to Cloudflare Queues. The queue
    # field is a StringValueOrRef flattened to a plain string by the tfvars converter.
    event_notifications = optional(list(object({
      queue = string
      rules = list(object({
        actions     = list(string)
        description = optional(string, "")
        prefix      = optional(string, "")
        suffix      = optional(string, "")
      }))
    })), [])
  })
}
