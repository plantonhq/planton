variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "KubernetesManifest specification"
  type = object({
    namespace = string
    create_namespace = optional(bool, false)
    manifest_yaml = string
    skip_await = optional(bool, false)
  })
}