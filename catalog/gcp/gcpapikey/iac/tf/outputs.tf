# The key's full resource name: projects/{project}/locations/global/keys/{key_id}.
# The provider's `name` attribute is the short key_id; the resource id is the
# handle the API Keys API addresses the key by.
output "name" {
  description = "The API key's full resource name"
  value       = google_apikeys_key.this.id
}

# The key's unique id -- what a Firebase app registration's api_key_id
# references.
output "uid" {
  description = "The API key's unique id"
  value       = google_apikeys_key.this.uid
}

# The key string the client presents. A credential Google bills against the
# project -- marked sensitive so it never prints in plans; the Pulumi module
# exports the same output as a secret.
output "key_string" {
  description = "The API key string"
  value       = google_apikeys_key.this.key_string
  sensitive   = true
}
