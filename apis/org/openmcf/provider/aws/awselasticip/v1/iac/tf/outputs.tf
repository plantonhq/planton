output "allocation_id" {
  description = "The allocation ID of the Elastic IP (e.g., eipalloc-xxx)."
  value       = aws_eip.this.allocation_id
}

output "public_ip" {
  description = "The public IPv4 address assigned to this Elastic IP."
  value       = aws_eip.this.public_ip
}

output "arn" {
  description = "The ARN of the Elastic IP."
  value       = aws_eip.this.arn
}

output "public_dns" {
  description = "The public DNS hostname associated with the Elastic IP."
  value       = aws_eip.this.public_dns
}
