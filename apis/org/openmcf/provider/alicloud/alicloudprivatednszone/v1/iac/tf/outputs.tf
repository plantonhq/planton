output "zone_id" {
  description = "The Private Zone ID assigned by Alibaba Cloud"
  value       = alicloud_pvtz_zone.main.id
}

output "zone_name" {
  description = "The zone name as created"
  value       = alicloud_pvtz_zone.main.zone_name
}

output "is_ptr" {
  description = "Whether the zone is a reverse-lookup (PTR) zone"
  value       = alicloud_pvtz_zone.main.is_ptr
}

output "record_count" {
  description = "The number of DNS records in the zone"
  value       = alicloud_pvtz_zone.main.record_count
}
