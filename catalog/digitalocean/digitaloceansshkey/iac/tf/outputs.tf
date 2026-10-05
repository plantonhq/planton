# Outputs — exactly the DigitalOceanSshKeyOutputs contract,
# identical across both provisioners.

output "ssh_key_id" {
  description = "Numeric id of the SSH key (the API identity and the only import id)"
  value       = digitalocean_ssh_key.ssh_key.id
}

output "fingerprint" {
  description = "MD5 fingerprint of the key, computed by DigitalOcean"
  value       = digitalocean_ssh_key.ssh_key.fingerprint
}
