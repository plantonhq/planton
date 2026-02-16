output "web_acl_arn" {
  description = "The ARN of the WAFv2 Web ACL."
  value       = aws_wafv2_web_acl.this.arn
}

output "web_acl_id" {
  description = "The unique identifier of the WAFv2 Web ACL."
  value       = aws_wafv2_web_acl.this.id
}

output "web_acl_name" {
  description = "The name of the WAFv2 Web ACL."
  value       = aws_wafv2_web_acl.this.name
}

output "capacity" {
  description = "The WCUs consumed by all rules in this Web ACL."
  value       = aws_wafv2_web_acl.this.capacity
}
