variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)."
  type        = any
}

variable "spec" {
  description = "AwsEventBridgeRuleSpec — desired configuration for the EventBridge rule and targets."
  type        = any
}
