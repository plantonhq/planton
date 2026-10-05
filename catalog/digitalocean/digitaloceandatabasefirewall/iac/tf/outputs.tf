# Outputs — exactly the DigitalOceanDatabaseFirewallOutputs
# contract, identical across both provisioners. The rule set is a property
# of its cluster; the cluster UUID is the only durable identity (the
# Terraform state id is a random unique string).

output "cluster_id" {
  description = "UUID of the database cluster whose inbound sources this rule set defines"
  value       = digitalocean_database_firewall.firewall.cluster_id
}
