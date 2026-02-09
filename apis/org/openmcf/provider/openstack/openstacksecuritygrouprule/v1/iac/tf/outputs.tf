# outputs.tf

output "rule_id" {
  description = "The UUID of the created security group rule"
  value       = openstack_networking_secgroup_rule_v2.main.id
}

output "security_group_id" {
  description = "The UUID of the parent security group"
  value       = openstack_networking_secgroup_rule_v2.main.security_group_id
}

output "direction" {
  description = "The direction of the rule (ingress or egress)"
  value       = openstack_networking_secgroup_rule_v2.main.direction
}

output "protocol" {
  description = "The IP protocol of the rule"
  value       = openstack_networking_secgroup_rule_v2.main.protocol
}

output "port_range_min" {
  description = "The lower bound of the port range (or ICMP type)"
  value       = openstack_networking_secgroup_rule_v2.main.port_range_min
}

output "port_range_max" {
  description = "The upper bound of the port range (or ICMP code)"
  value       = openstack_networking_secgroup_rule_v2.main.port_range_max
}

output "region" {
  description = "The OpenStack region where the rule was created"
  value       = openstack_networking_secgroup_rule_v2.main.region
}
