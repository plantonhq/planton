variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)."
  type        = any
}

variable "spec" {
  description = "AwsCloudwatchLogGroupSpec — desired configuration for the CloudWatch Log Group."
  type        = any
}
