# The R2 key pair that reads spec.r2_bundle through R2's S3-compatible API.
#
# R2 is a separate credential plane from the Cloudflare API token: an S3 request is
# signed with an access key pair, which the API token cannot do. Planton delivers the
# pair from the Cloudflare connection's r2 block as TF_VAR_r2_access_key_id /
# TF_VAR_r2_secret_access_key (and TF_VAR_r2_endpoint when the connection overrides the
# endpoint) -- the same pair the Pulumi module reads from the provider config. The pair
# never rides AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY, because an S3 state backend
# authenticated from the environment reads those same names.
#
# All three default to null. With a bundle and a null key, the storage provider falls
# back to its own credential chain (AWS_* in a developer shell), exactly as the Pulumi
# module does without an r2 block. Without a bundle these variables are never read;
# see provider.tf for why the storage provider still configures cleanly then.
variable "r2_access_key_id" {
  description = "R2 access key ID that can read the bundle bucket; null defers to the storage provider's own credential chain"
  type        = string
  default     = null
  sensitive   = true
}

variable "r2_secret_access_key" {
  description = "R2 secret access key paired with r2_access_key_id"
  type        = string
  default     = null
  sensitive   = true
}

variable "r2_endpoint" {
  description = "R2 S3-compatible endpoint; null derives https://<account_id>.r2.cloudflarestorage.com from spec.account_id"
  type        = string
  default     = null
}
