output "internet_gateway_id" {
  description = "The internet gateway's id"
  value       = aws_internet_gateway.this.id
}

output "internet_gateway_arn" {
  description = "The internet gateway's ARN"
  value       = aws_internet_gateway.this.arn
}

output "vpc_id" {
  description = "The id of the VPC this gateway is attached to"
  value       = var.spec.vpc_id
}

output "region" {
  description = "The AWS region the internet gateway was created in"
  value       = var.spec.region
}
