# outputs.tf

output "security_group_id" {
  description = "The ID of the newly created Security Group."
  value       = aws_security_group.this.id
}

