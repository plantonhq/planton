# Load Balancer ID
output "lb_id" {
  description = "The unique identifier of the created Load Balancer"
  value       = scaleway_lb.lb.id
}

# Public IP Address
output "lb_ip_address" {
  description = "The public IPv4 address assigned to the Load Balancer"
  value       = scaleway_lb_ip.ip.ip_address
}

# Flexible IP ID
output "lb_ip_id" {
  description = "The unique identifier of the Flexible IP resource"
  value       = scaleway_lb_ip.ip.id
}
