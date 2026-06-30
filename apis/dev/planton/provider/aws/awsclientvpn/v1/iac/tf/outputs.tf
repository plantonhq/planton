output "client_vpn_endpoint_id" {
  value       = aws_ec2_client_vpn_endpoint.this.id
  description = "The AWS-assigned identifier for the Client VPN endpoint."
}

output "security_group_id" {
  value       = length(aws_security_group.this) > 0 ? aws_security_group.this[0].id : null
  description = "Security group ID applied to the endpoint associations."
}

output "subnet_association_ids" {
  value       = { for k, v in aws_ec2_client_vpn_network_association.this : k => v.id }
  description = "Map of subnet ID to association ID."
}

output "endpoint_dns_name" {
  value       = aws_ec2_client_vpn_endpoint.this.dns_name
  description = "The DNS name clients use to connect to the Client VPN endpoint."
}


