# outputs.tf

output "name" {
  description = "The name of the keypair"
  value       = openstack_compute_keypair_v2.main.name
}

output "fingerprint" {
  description = "The MD5 fingerprint of the SSH public key"
  value       = openstack_compute_keypair_v2.main.fingerprint
}

output "public_key" {
  description = "The SSH public key in OpenSSH authorized_keys format"
  value       = openstack_compute_keypair_v2.main.public_key
}

output "region" {
  description = "The OpenStack region where the keypair was created"
  value       = openstack_compute_keypair_v2.main.region
}

# The private key is only available when OpenStack generates the keypair
# (i.e., when no public_key is provided in the spec).
# It is marked sensitive to prevent it from appearing in logs.
output "private_key" {
  description = "The generated private key (only available when no public_key is provided)"
  value       = openstack_compute_keypair_v2.main.private_key
  sensitive   = true
}
