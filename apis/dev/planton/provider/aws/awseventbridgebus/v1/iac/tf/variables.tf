variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)."
  type        = any
}

variable "spec" {
  description = "AwsEventBridgeBusSpec — desired configuration for the EventBridge bus."
  type        = any
}
