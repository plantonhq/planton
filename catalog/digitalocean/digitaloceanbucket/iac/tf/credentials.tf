# The API token. When null (the default), the provider falls back to its own
# environment defaults (DIGITALOCEAN_TOKEN, then DIGITALOCEAN_ACCESS_TOKEN) --
# the same fallback the Pulumi module gets from the bridge, so both engines
# authenticate identically whether the token arrives as TF_VAR_digitalocean_token
# (the platform's stack-input bridge) or as ambient environment (a developer
# shell, the E2E harness). A required variable here would fail every apply that
# relies on the environment, which is the provider's own documented contract.
variable "digitalocean_token" {
  description = "DigitalOcean API token for authentication; null defers to DIGITALOCEAN_TOKEN / DIGITALOCEAN_ACCESS_TOKEN"
  type        = string
  default     = null
  sensitive   = true
}

# Spaces is an S3-compatible credential plane the API token cannot reach.
# When null, the provider falls back to its own env defaults
# (SPACES_ACCESS_KEY_ID / SPACES_SECRET_ACCESS_KEY).
variable "spaces_access_id" {
  description = "DigitalOcean Spaces access key id"
  type        = string
  default     = null
  sensitive   = true
}

variable "spaces_secret_key" {
  description = "DigitalOcean Spaces secret access key"
  type        = string
  default     = null
  sensitive   = true
}
