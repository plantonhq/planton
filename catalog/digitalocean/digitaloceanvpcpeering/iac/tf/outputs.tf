# Stack outputs — exactly the DigitalOceanVpcPeeringStackOutputs contract,
# identical across both provisioners.
#
# The lifecycle status is deliberately not exported: the provider waits for
# ACTIVE before the apply succeeds, so a stored status could only ever read
# "ACTIVE" and would go stale the moment DigitalOcean moved the peering.
# Live status belongs to whoever reads the API.

output "peering_id" {
  description = "UUID of the VPC peering connection (its API identity and import id)"
  value       = digitalocean_vpc_peering.peering.id
}
