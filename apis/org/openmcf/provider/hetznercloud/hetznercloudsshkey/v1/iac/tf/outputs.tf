output "ssh_key_id" {
  description = "The Hetzner Cloud numeric ID of the created SSH key"
  value       = hcloud_ssh_key.this.id
}

output "fingerprint" {
  description = "MD5 fingerprint of the SSH public key"
  value       = hcloud_ssh_key.this.fingerprint
}
