variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)."
  type        = any
}

variable "spec" {
  description = "AwsStepFunctionSpec — desired configuration for the state machine."
  type        = any
}
