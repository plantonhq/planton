variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)."
  type        = any
}

variable "spec" {
  description = "AwsMemcachedElasticacheSpec — desired configuration for the ElastiCache Memcached cluster."
  type        = any
}
