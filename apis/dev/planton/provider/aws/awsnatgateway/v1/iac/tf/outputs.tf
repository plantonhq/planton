output "nat_gateway_id" {
  description = "The NAT gateway's id"
  value       = aws_nat_gateway.this.id
}

output "public_ip" {
  description = "The public IPv4 address of a public gateway (empty for a private gateway)"
  value       = aws_nat_gateway.this.public_ip
}

output "private_ip" {
  description = "The private IPv4 address assigned to the gateway within its subnet"
  value       = aws_nat_gateway.this.private_ip
}

output "network_interface_id" {
  description = "The id of the elastic network interface AWS created for the gateway"
  value       = aws_nat_gateway.this.network_interface_id
}

output "subnet_id" {
  description = "The id of the subnet the gateway lives in"
  value       = var.spec.subnet_id
}

output "region" {
  description = "The AWS region the NAT gateway was created in"
  value       = var.spec.region
}
