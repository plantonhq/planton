variable "metadata" {
  description = "metadata for all resource objects on planton"
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
    labels = object({

      # Description for key
      key = string

      # Description for value
      value = string
    })

    # annotations for the resource
    annotations = object({

      # Description for key
      key = string

      # Description for value
      value = string
    })

    # tags for the resource
    tags = list(string)
  })
}

variable "spec" {
  description = "Specification for Deployment Component"
  type = object({

    # The AWS region where the resource will be created.
    region = string

    # Description for user_name
    user_name = string

    # Description for managed_policy_arns
    managed_policy_arns = list(string)

    # Description for inline_policies
    inline_policies = object({

      # Description for key
      key = string

      # Description for value
      value = string
    })

    # Description for disable_access_keys
    disable_access_keys = bool
  })
}