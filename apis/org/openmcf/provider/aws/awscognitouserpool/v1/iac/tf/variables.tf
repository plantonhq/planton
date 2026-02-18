variable "metadata" {
  description = "Resource metadata (name, org, env, id)"
  type = object({
    name = string
    org  = string
    env  = string
    id   = string
  })
}

variable "spec" {
  description = "AwsCognitoUserPoolSpec configuration"
  type = object({
    # The AWS region where the resource will be created.
    region = string
  })
}
