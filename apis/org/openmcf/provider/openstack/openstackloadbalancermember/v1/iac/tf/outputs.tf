# outputs.tf

output "member_id" {
  description = "The unique identifier (UUID) of the member"
  value       = openstack_lb_member_v2.main.id
}

output "name" {
  description = "The name of the member"
  value       = openstack_lb_member_v2.main.name
}

output "address" {
  description = "The IP address of the backend server"
  value       = openstack_lb_member_v2.main.address
}

output "protocol_port" {
  description = "The port on the backend server"
  value       = openstack_lb_member_v2.main.protocol_port
}

output "weight" {
  description = "The member weight in the load-balancing algorithm"
  value       = openstack_lb_member_v2.main.weight
}

output "region" {
  description = "The OpenStack region where the member was created"
  value       = openstack_lb_member_v2.main.region
}
