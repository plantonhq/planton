output "bucket_id" {
  description = "ID (name) of the S3 bucket created on AWS"
  value       = aws_s3_bucket.this.id
}

output "bucket_arn" {
  description = "ARN (Amazon Resource Name) of the S3 bucket"
  value       = aws_s3_bucket.this.arn
}

output "region" {
  description = "AWS region where the bucket is created"
  value       = aws_s3_bucket.this.region
}

output "bucket_regional_domain_name" {
  description = "Regional domain name of the S3 bucket"
  value       = aws_s3_bucket.this.bucket_regional_domain_name
}

output "hosted_zone_id" {
  description = "Hosted zone ID for the S3 bucket's region"
  value       = aws_s3_bucket.this.hosted_zone_id
}
