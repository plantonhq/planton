variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)."
  type        = any
}

variable "spec" {
  description = "AwsServerlessElasticacheSpec — desired configuration for the ElastiCache Serverless cache."
  type        = any
}
