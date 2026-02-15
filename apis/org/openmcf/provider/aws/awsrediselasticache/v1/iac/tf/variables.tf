variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)."
  type        = any
}

variable "spec" {
  description = "AwsRedisElasticacheSpec — desired configuration for the ElastiCache Redis/Valkey cluster."
  type        = any
}
