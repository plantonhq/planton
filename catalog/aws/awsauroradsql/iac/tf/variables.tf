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
  description = "AwsAuroraDsql specification"
  type = object({
    region = string
    deletion_protection_enabled = optional(bool, false)
    force_destroy = optional(bool, false)
    kms_encryption_key = optional(string, "")
    multi_region = optional(object({
      witness_region = string
      peer_cluster_arns = list(string)
    }))
  })
}