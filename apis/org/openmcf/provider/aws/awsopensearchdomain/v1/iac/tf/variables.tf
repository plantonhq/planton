variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)."
  type        = any
}

variable "spec" {
  description = "AwsOpenSearchDomainSpec — desired configuration for the OpenSearch domain."
  type        = any
}
