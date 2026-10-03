# Pulumi Module: DigitalOcean CDN

Provisions a CDN endpoint fronting a Spaces bucket -- the complete `digitalocean_cdn` resource surface at schema v1. Behavioral parity with the Terraform module is the contract.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean.Cdn` | The edge endpoint: origin + ttl + certificate name + custom domain |

## Inputs

`DigitalOceanCdnIacInput`: the target `DigitalOceanCdn` resource and the DigitalOcean provider config (API token).

## Outputs

Exactly the `DigitalOceanCdnOutputs` contract: `cdn_id` (Pulumi's resource id), `endpoint`.

## Behavior notes

- The certificate is wired through the SDK's `CertificateName` input ONLY: the deprecated `CertificateId` is never set (its update path silently detaches the certificate when the custom domain changes).
- Unset `ttl` stays nil and defers to DigitalOcean's 3600 default; presence-gated from the spec's optional field.
- `origin` is create-only: changing it replaces the endpoint (and its hostname).
