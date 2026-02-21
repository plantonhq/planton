output "domain_name" {
  description = "The accelerated domain name as registered in CDN"
  value       = alicloud_cdn_domain_new.main.domain_name
}

output "cname" {
  description = "The CNAME assigned by Alibaba Cloud CDN"
  value       = alicloud_cdn_domain_new.main.cname
}

output "status" {
  description = "The current status of the CDN domain"
  value       = alicloud_cdn_domain_new.main.status
}
