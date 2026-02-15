variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)."
  type        = any
}

variable "spec" {
  description = "AwsSnsTopicSpec — desired configuration for the SNS topic."
  type        = any
}
