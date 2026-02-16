output "domain_id" {
  description = "The unique identifier of the OpenSearch domain."
  value       = aws_opensearch_domain.this.domain_id
}

output "domain_name" {
  description = "The name of the OpenSearch domain."
  value       = aws_opensearch_domain.this.domain_name
}

output "domain_arn" {
  description = "The ARN of the OpenSearch domain."
  value       = aws_opensearch_domain.this.arn
}

output "endpoint" {
  description = "The domain-specific endpoint for index and search requests."
  value       = aws_opensearch_domain.this.endpoint
}

output "dashboard_endpoint" {
  description = "The OpenSearch Dashboards endpoint."
  value       = aws_opensearch_domain.this.dashboard_endpoint
}
