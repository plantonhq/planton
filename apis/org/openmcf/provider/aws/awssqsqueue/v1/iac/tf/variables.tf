variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)."
  type        = any
}

variable "spec" {
  description = "AwsSqsQueueSpec — desired configuration for the SQS queue."
  type        = any
}
