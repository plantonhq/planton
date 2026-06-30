variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)."
  type        = any
}

variable "spec" {
  description = "AwsKinesisStreamConsumerSpec — desired configuration for the Kinesis stream consumer."
  type        = any
}
