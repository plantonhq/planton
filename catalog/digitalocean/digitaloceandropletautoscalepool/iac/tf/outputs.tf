# Stack outputs — exactly the DigitalOceanDropletAutoscalePoolStackOutputs
# contract, identical across both provisioners. The pool's health is
# deliberately not an output: an apply-time status goes stale the moment
# DigitalOcean changes it, so live health is read from the API, never from
# stored outputs.

output "pool_id" {
  description = "UUID of the autoscale pool (its API identity and import id)"
  value       = digitalocean_droplet_autoscale.pool.id
}
