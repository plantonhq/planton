output "domain_id" {
  description = "The domain ID assigned by Alibaba Cloud"
  value       = alicloud_alidns_domain.main.domain_id
}

output "domain_name" {
  description = "The domain name as registered in Alidns"
  value       = alicloud_alidns_domain.main.domain_name
}

output "dns_servers" {
  description = "DNS server names assigned by Alibaba Cloud"
  value       = tolist(alicloud_alidns_domain.main.dns_servers)
}

output "group_name" {
  description = "The domain group name (computed from group_id)"
  value       = alicloud_alidns_domain.main.group_name
}

output "puny_code" {
  description = "Punycode representation for internationalized domain names"
  value       = alicloud_alidns_domain.main.puny_code
}
