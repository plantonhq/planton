output "bucket_id" {
  description = "The name (id) of the S3 bucket."
  value       = aws_s3_bucket.this.id
}

output "bucket_arn" {
  description = "The ARN of the S3 bucket (arn:aws:s3:::bucket-name)."
  value       = aws_s3_bucket.this.arn
}

output "region" {
  description = "The AWS region the bucket lives in."
  value       = aws_s3_bucket.this.region
}

output "bucket_regional_domain_name" {
  description = "Regional domain name (bucket-name.s3.region.amazonaws.com)."
  value       = aws_s3_bucket.this.bucket_regional_domain_name
}

output "hosted_zone_id" {
  description = "Route53 hosted zone ID of the bucket's region, for alias records."
  value       = aws_s3_bucket.this.hosted_zone_id
}

output "bucket_domain_name" {
  description = "Global domain name (bucket-name.s3.amazonaws.com)."
  value       = aws_s3_bucket.this.bucket_domain_name
}

# Website outputs come from the website satellite and are empty when static
# website hosting is not configured — stated as empty strings (not null) so
# the output contract is shape-stable across both engines.
output "website_endpoint" {
  description = "S3 website endpoint, populated only when website hosting is configured."
  value       = length(aws_s3_bucket_website_configuration.this) > 0 ? aws_s3_bucket_website_configuration.this[0].website_endpoint : ""
}

output "website_domain" {
  description = "S3 website service domain for Route53 aliases, populated only when website hosting is configured."
  value       = length(aws_s3_bucket_website_configuration.this) > 0 ? aws_s3_bucket_website_configuration.this[0].website_domain : ""
}
