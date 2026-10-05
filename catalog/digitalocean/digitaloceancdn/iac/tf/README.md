# Terraform Module: DigitalOcean CDN

Provisions a CDN endpoint fronting a Spaces bucket -- the complete `digitalocean_cdn` resource surface at schema v1.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean_cdn.cdn` | The edge endpoint: origin + ttl + certificate name + custom domain |

## Inputs

Generated `variables.tf` mirrors the `DigitalOceanCdnSpec` proto: the flattened `origin` reference string (the Space's FQDN), presence-typed `ttl`, the flattened `certificate` reference string (the certificate NAME), and `custom_domain`. Authentication uses `digitalocean_token` (sensitive).

## Outputs

Exactly the `DigitalOceanCdnOutputs` contract: `cdn_id`, `endpoint`.

## Behavior notes

- The certificate is wired through `certificate_name` ONLY: the deprecated `certificate_id` argument is never rendered (its update path silently detaches the certificate when the custom domain changes).
- Unset `ttl` defers to DigitalOcean's 3600 default, read back without a perpetual diff (Optional+Computed); an explicit zero is unrepresentable (spec floor 1 -- the provider drops zeros on the wire).
- `origin` is create-only: changing it replaces the endpoint (and its hostname).
- The provider retries reads on 404 for 30s (edge creation is eventually consistent) -- and its read-after-destroy ERRORS instead of settling; the E2E verifier owns destroy proof.
- Import: `terraform import ... <cdn_id>` (see `iac/import-map.yaml`).
