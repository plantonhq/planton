variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)."
  type        = any
}

variable "spec" {
  description = "AwsElasticIpSpec — desired configuration for the Elastic IP."
  type        = any
}
