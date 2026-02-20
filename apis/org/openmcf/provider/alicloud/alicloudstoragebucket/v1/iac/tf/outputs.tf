output "bucket_name" {
  description = "The bucket name (also the bucket ID)"
  value       = alicloud_oss_bucket.main.bucket
}

output "extranet_endpoint" {
  description = "The public internet endpoint for the bucket"
  value       = alicloud_oss_bucket.main.extranet_endpoint
}

output "intranet_endpoint" {
  description = "The VPC-internal endpoint for the bucket"
  value       = alicloud_oss_bucket.main.intranet_endpoint
}
