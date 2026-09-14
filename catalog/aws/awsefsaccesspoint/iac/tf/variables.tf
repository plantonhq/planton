variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "AwsEfsAccessPoint specification"
  type = object({
    region = string
    file_system_id = string
    posix_user = optional(object({
      uid = number
      gid = number
      secondary_gids = optional(list(number), [])
    }))
    root_directory = optional(object({
      path = string
      creation_info = optional(object({
        owner_uid = optional(number, 0)
        owner_gid = optional(number, 0)
        permissions = string
      }))
    }))
  })
}
