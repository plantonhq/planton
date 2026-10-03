# DigitalOcean CDN

Built for 100% parity with the Terraform DigitalOcean provider's `digitalocean_cdn` resource at the pinned provider version.

## What this kind models

A CDN endpoint that serves a Spaces bucket's content from DigitalOcean's global edge network -- optionally under your own subdomain with a managed TLS certificate.

- `origin` -- the Space's FQDN, wired by reference to a DigitalOceanBucket (its `bucket_domain_name` output); create-only
- `ttl` -- edge cache seconds (unset = DigitalOcean's 3600 default; the floor is 1 because an explicit zero can never reach the API)
- `certificate` -- the TLS certificate for a custom domain, referenced by its stable NAME (exactly what DigitalOceanCertificate exports as `certificate_id`)
- `custom_domain` -- your subdomain (requires `certificate`, enforced at validation)

## Quick start

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanCdn
metadata:
  name: app-assets-cdn
spec:
  origin:
    valueFrom:
      kind: DigitalOceanBucket
      name: app-assets
      fieldPath: status.outputs.bucket_domain_name
```

Deploy with either provisioner; both produce identical resources and outputs.

## Outputs

| Output | Description |
|---|---|
| `cdn_id` | UUID of the CDN endpoint (its API identity and import id) |
| `endpoint` | The FQDN the CDN serves content from -- point custom-domain CNAMEs here |

## Behavior worth knowing

- **Certificates are referenced by NAME, never UUID.** Let's Encrypt renewals rotate a certificate's UUID while the name stays stable; the provider's deprecated numeric `certificate_id` argument is deliberately not modeled -- its update path can silently DETACH the certificate when the custom domain changes.
- **Edge creation is eventually consistent.** The provider retries reads on 404 for up to 30 seconds after create.
- **The literal `needs-cloudflare-cert`** is a DigitalOcean sentinel value for the certificate reference, passed through verbatim.

## Module layout

- `iac/tf/` -- OpenTofu/Terraform module (provider pinned `~> 2.99`)
- `iac/pulumi/` -- Pulumi module (Go, pulumi-digitalocean SDK)
- Both engines wire the same spec fields and export the same outputs; behavioral parity is the contract.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
