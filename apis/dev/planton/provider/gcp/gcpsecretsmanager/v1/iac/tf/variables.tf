variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name = string,
    id = optional(string),
    org = optional(string),
    env = optional(string),
    labels = optional(map(string)),
    tags = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}


variable "spec" {
  description = "spec"
  type = object({

    # The GCP project ID where the secrets will be created.
    # Can be provided as a literal value or as a reference to another resource's output.
    # Example (literal): {value = "my-gcp-project-123456"}
    # Example (reference): {value_from = {kind = "GcpProject", name = "main-project"}}
    project_id = object({
      value      = optional(string)
      value_from = optional(object({
        kind       = optional(string)
        env        = optional(string)
        name       = string
        field_path = optional(string)
      }))
    })

    # A list of secret names to create in Google Cloud Secrets Manager.
    # Each name represents a unique secret that can store sensitive data securely.
    secret_names = list(string)
  })
}
