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
  description = "AwsS3DirectoryBucket specification"
  type = object({
    region = string
    zone_id = string
    zone_type = optional(string, "")
    data_redundancy = optional(string, "")
    force_destroy = optional(bool, false)
  })
}