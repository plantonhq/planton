variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)."
  type        = any
}

variable "spec" {
  description = "AwsKinesisFirehoseSpec — desired configuration for the Firehose delivery stream."
  type        = any
}
