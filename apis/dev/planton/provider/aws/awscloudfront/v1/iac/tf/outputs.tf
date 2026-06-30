output "distribution_id" {
  value       = aws_cloudfront_distribution.this.id
  description = "Distribution ID (e.g., E123ABCXYZ)."
}

output "domain_name" {
  value       = aws_cloudfront_distribution.this.domain_name
  description = "CloudFront distribution domain name (e.g., d123.cloudfront.net)."
}

output "hosted_zone_id" {
  value       = "Z2FDTNDATAQYW2"
  description = "Route53 hosted zone ID for aliasing to CloudFront."
}


