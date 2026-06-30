variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)."
  type        = any
}

variable "spec" {
  description = "AwsKinesisStreamSpec — desired configuration for the Kinesis stream."
  type        = any
}
