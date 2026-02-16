resource "aws_kinesis_stream_consumer" "this" {
  name       = local.consumer_name
  stream_arn = local.stream_arn
}
