# Outputs for the DigitalOcean Load Balancer

output "load_balancer_id" {
  description = "The unique identifier (UUID) of the load balancer"
  value       = digitalocean_loadbalancer.this.id
}

output "load_balancer_ip" {
  description = "The public IP address of the load balancer"
  value       = digitalocean_loadbalancer.this.ip
}

output "load_balancer_urn" {
  description = "The uniform resource name (URN) of the load balancer"
  value       = digitalocean_loadbalancer.this.urn
}

output "load_balancer_name" {
  description = "The name of the load balancer"
  value       = digitalocean_loadbalancer.this.name
}

output "status" {
  description = "The status of the load balancer (new, active, or errored)"
  value       = digitalocean_loadbalancer.this.status
}

# Complete outputs object matching stack_outputs.proto structure
output "outputs" {
  description = "Complete load balancer outputs for integration with other resources"
  value = {
    load_balancer_id = digitalocean_loadbalancer.this.id
    ip               = digitalocean_loadbalancer.this.ip
    dns_name         = digitalocean_loadbalancer.this.name  # DigitalOcean LBs don't have DNS names, use name as placeholder
  }
}

