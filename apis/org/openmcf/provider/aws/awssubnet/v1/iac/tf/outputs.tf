output "subnet_id" {
  description = "The subnet's id"
  value       = aws_subnet.this.id
}

output "subnet_arn" {
  description = "The subnet's ARN"
  value       = aws_subnet.this.arn
}

output "availability_zone" {
  description = "The availability zone the subnet resides in"
  value       = aws_subnet.this.availability_zone
}

output "cidr_block" {
  description = "The subnet's IPv4 CIDR block"
  value       = aws_subnet.this.cidr_block
}

output "route_table_id" {
  description = "The route table associated with this subnet (inline-created, externally referenced, or empty when on the VPC main route table)"
  value = (
    length(var.spec.routes) > 0
    ? aws_route_table.this[0].id
    : var.spec.route_table_id
  )
}

output "region" {
  description = "The AWS region the subnet was created in"
  value       = var.spec.region
}
