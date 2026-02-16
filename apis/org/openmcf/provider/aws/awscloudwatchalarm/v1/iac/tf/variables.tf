variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)."
  type        = any
}

variable "spec" {
  description = "AwsCloudwatchAlarmSpec — desired configuration for the CloudWatch Alarm."
  type        = any
}
