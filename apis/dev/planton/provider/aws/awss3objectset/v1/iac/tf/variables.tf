variable "metadata" {
  description = "metadata"
  type = object({
    # name of the resource
    name = string
    # id of the resource
    id = string
    # id of the organization to which the api-resource belongs to
    org = string
    # environment to which the resource belongs to
    env = string
    # labels for the resource
    labels = map(string)
    # annotations for the resource
    annotations = map(string)
    # tags for the resource
    tags = list(string)
  })
}

variable "spec" {
  description = "spec"
  type = object({
    # The target S3 bucket name (resolved from foreign key)
    bucket = string

    # The AWS region where the S3 bucket is located.
    region = string

    # Tags applied to all objects in the set
    tags = optional(map(string), {})

    # List of S3 objects to upload
    objects = list(object({
      # The S3 object key (path within the bucket)
      key = string

      # Inline UTF-8 text content (mutually exclusive with content_base64)
      content = optional(string, null)

      # Base64-encoded binary content (mutually exclusive with content)
      content_base64 = optional(string, null)

      # MIME content type (e.g., "application/json", "text/html")
      content_type = optional(string, null)

      # Cache-Control header value
      cache_control = optional(string, null)

      # Content-Encoding header value (e.g., "gzip")
      content_encoding = optional(string, null)

      # Per-object tags (merged with set-level tags)
      tags = optional(map(string), {})

      # Canned ACL for the object
      acl = optional(string, null)
    }))
  })
}
