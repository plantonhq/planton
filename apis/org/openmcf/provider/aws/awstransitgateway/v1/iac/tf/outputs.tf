output "transit_gateway_id" {
  description = "The Transit Gateway ID"
  value       = aws_ec2_transit_gateway.this.id
}

output "transit_gateway_arn" {
  description = "The Transit Gateway ARN"
  value       = aws_ec2_transit_gateway.this.arn
}

output "owner_id" {
  description = "The AWS account ID that owns the Transit Gateway"
  value       = aws_ec2_transit_gateway.this.owner_id
}

output "association_default_route_table_id" {
  description = "The ID of the default association route table"
  value       = aws_ec2_transit_gateway.this.association_default_route_table_id
}

output "propagation_default_route_table_id" {
  description = "The ID of the default propagation route table"
  value       = aws_ec2_transit_gateway.this.propagation_default_route_table_id
}

output "vpc_attachment_ids" {
  description = "Map of attachment name to VPC attachment ID"
  value       = { for name, att in aws_ec2_transit_gateway_vpc_attachment.this : name => att.id }
}
