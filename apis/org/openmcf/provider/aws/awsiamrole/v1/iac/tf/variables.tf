variable "metadata" {
  description = "metadata captures identifying information (name, org, version, etc.)\nand must pass standard validations for resource naming."
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
  description = "spec holds the core configuration data defining how the ECS service is deployed."
  type = object({

    # The AWS region where the resource will be created.
    region = string

    # description is an optional description of the IAM role.
    description = string

    # path is the IAM path for the role. Defaults to "/" if omitted.
    path = string

    # trust_policy_json is the JSON string describing the trust relationship for the role.
    # Example: a trust policy allowing ECS tasks to assume this role.
    trust_policy = string

    # managed_policy_arns is a list of ARNs for AWS-managed or customer-managed IAM policies
    # you want to attach to this role.
    managed_policy_arns = list(string)

    # inline_policy_jsons is a map of inline policy names to a JSON policy doc.
    # Key is policy name. Value is the raw JSON for that policy.
    inline_policies = object({

      # Description for key
      key = string

      # Description for value
      value = string
    })
  })
}