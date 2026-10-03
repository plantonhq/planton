# Outputs — exactly the DigitalOceanLoadBalancerOutputs
# contract, identical across both provisioners.

output "load_balancer_id" {
  description = "The unique identifier (UUID) of the load balancer"
  value       = digitalocean_loadbalancer.this.id
}

output "ip" {
  description = "The public IPv4 address assigned to the load balancer"
  value       = digitalocean_loadbalancer.this.ip
}

output "urn" {
  description = "The uniform resource name of the load balancer (do:loadbalancer:<id>)"
  value       = digitalocean_loadbalancer.this.urn
}

output "ipv6" {
  description = "The IPv6 address of the load balancer (populated when network_stack is DUALSTACK)"
  value       = digitalocean_loadbalancer.this.ipv6
}
