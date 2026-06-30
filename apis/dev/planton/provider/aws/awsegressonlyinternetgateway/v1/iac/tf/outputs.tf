output "egress_only_internet_gateway_id" {
  description = "The egress-only internet gateway's id"
  value       = aws_egress_only_internet_gateway.this.id
}

output "vpc_id" {
  description = "The id of the VPC this gateway is attached to"
  value       = var.spec.vpc_id
}

output "region" {
  description = "The AWS region the egress-only internet gateway was created in"
  value       = var.spec.region
}
