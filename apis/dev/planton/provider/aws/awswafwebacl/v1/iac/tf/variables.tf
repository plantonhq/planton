variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)."
  type        = any
}

variable "spec" {
  description = "AwsWafWebAclSpec — desired configuration for the WAFv2 Web ACL."
  type        = any
}
