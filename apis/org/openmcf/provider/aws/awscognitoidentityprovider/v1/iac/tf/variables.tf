variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)."
  type        = any
}

variable "spec" {
  description = "AwsCognitoIdentityProviderSpec — desired configuration for the identity provider."
  type        = any
}
