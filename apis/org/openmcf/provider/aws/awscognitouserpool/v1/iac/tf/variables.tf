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
  type        = any
}
