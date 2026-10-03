# DigitalOcean CDN

Puts a Spaces bucket behind DigitalOcean's global edge network, optionally under a custom subdomain with a managed TLS certificate. The endpoint itself is free -- CDN delivery is included with the Spaces subscription, and edge bandwidth draws from the origin bucket's existing transfer allowance. The origin is create-only: changing it replaces the endpoint and its hostname, so front the endpoint with your own custom domain early if consumers will hard-code URLs.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **CDN Endpoint** -- the edge distribution fronting your referenced bucket, with your cache TTL and (optionally) your custom domain and certificate; the certificate is wired by its stable NAME, never by the deprecated numeric certificate id

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.
- **A bucket** -- the DigitalOceanBucket whose content the CDN serves, referenced (or named by its Spaces FQDN) in `origin`.
- **A certificate (optional)** -- a DigitalOceanCertificate, required only with `customDomain`.

### DigitalOcean Account

- **Publicly readable objects** -- the edge serves what the bucket allows; private objects return errors at the edge exactly as they do at the origin. Nothing else meters on this resource: bandwidth draws from the bucket's existing transfer allowance.

## Deploy

### Console

Open the deployment store, find **DigitalOcean CDN**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Public Assets CDN** preset in the [Presets](#presets) tab to front an assets bucket with a one-day cache TTL.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanCdn
metadata:
  name: app-assets-cdn
  org: acme-corp
  env: prod
spec:
  origin:
    value: app-assets.nyc3.digitaloceanspaces.com
  ttl: 86400
```

```shell
planton apply -f do-cdn.yaml
```

This fronts the bucket with DigitalOcean's edge network and a one-day cache TTL; the edge hostname lands in the `endpoint` output. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the origin bucket and the certificate deployed in the same InfraPipeline:

```yaml
spec:
  origin:
    valueFrom:
      kind: DigitalOceanBucket
      name: app-assets
      fieldPath: status.outputs.bucket_domain_name
  certificate:
    valueFrom:
      kind: DigitalOceanCertificate
      name: assets-cert
      fieldPath: status.outputs.certificate_id
  customDomain: assets.example.com
```

The InfraPipeline resolves the dependency graph, deploys the bucket and certificate first, then provisions the endpoint with the resolved values.

## Key Configuration

These are the most important decisions when configuring a CDN endpoint. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**`origin` is create-only -- and the endpoint hostname is an API** -- Changing the origin replaces the CDN endpoint, which changes the `endpoint` hostname; every CNAME and hard-coded URL pointing at the old one goes stale. Treat the hostname like an API surface: front it with your own `customDomain` early if consumers will hard-code URLs, so a future origin swap only re-points your CNAME.

**`ttl` is a blast-radius decision** -- The cache TTL applies endpoint-wide; DigitalOcean's CDN has no per-path rules. A long TTL (the 3600-second default and up) is right for fingerprinted assets that never change in place; content updated in place needs a short TTL and tolerance for staleness windows. An explicit zero is unrepresentable -- this is a cache, not a pass-through.

**`customDomain` requires `certificate`** -- DigitalOcean does not serve a custom domain without TLS, and the provider only surfaces that failure at apply time; the spec rejects the combination at validation time instead. The certificate is referenced by its stable NAME -- the same value DigitalOceanCertificate exports as `certificate_id` -- because every Let's Encrypt renewal mints a new certificate UUID while the name stays put. The provider's deprecated numeric certificate id argument can silently detach the certificate when the custom domain changes, and is deliberately unrepresentable here. The literal value `needs-cloudflare-cert` is a DigitalOcean sentinel requesting an edge-provisioned certificate instead of an account-managed one.

**The edge serves what the bucket allows** -- The CDN does not bypass bucket permissions. If content 403s at the edge, the fix is on the bucket (public-read objects or a bucket policy granting read), not here.

**Destroy is a traffic event** -- Tearing down the endpoint stops edge delivery immediately: the hostname stops serving and custom-domain CNAMEs dangle. The origin bucket and its content are untouched. Re-point DNS before destroying when a custom domain fronts real traffic.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanBucket** | `origin` | `status.outputs.bucket_domain_name` |
| **DigitalOceanCertificate** (optional, with `customDomain`) | `certificate` | `status.outputs.certificate_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `endpoint` | The fully-qualified hostname the CDN serves from (e.g. `my-bucket.nyc3.cdn.digitaloceanspaces.com`) | CNAME record targets for custom domains, application asset base URLs |
| `cdn_id` | UUID of the endpoint -- its API identity and import id | API operations, imports |

Serve content from `endpoint` instead of the bucket's own domain -- that is the hostname the edge network answers on.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Fingerprinted build artifacts** -- front an assets bucket with a one-day TTL: hashed JS/CSS bundles and versioned images never change in place, so long edge caching costs nothing in staleness. Start from the **Public Assets CDN** preset.

**Branded domain over TLS** -- the endpoint plus a managed certificate and a custom subdomain, so users see `assets.example.com` instead of a digitaloceanspaces.com host -- and the endpoint hostname stops being something consumers ever hard-code. Start from the **Branded Domain CDN** preset.

## Works With

- [**DigitalOcean Spaces Bucket**](/infra-catalog/digital-ocean-bucket) -- the origin whose content the edge serves; its permissions decide what the CDN can deliver
- [**DigitalOcean Certificate**](/infra-catalog/digital-ocean-certificate) -- the managed TLS certificate a custom domain requires, referenced by its stable name
- [**DigitalOcean DNS Record**](/infra-catalog/digital-ocean-dns-record) -- the CNAME pointing your custom domain at the `endpoint` output
- [**DigitalOcean DNS Zone**](/infra-catalog/digital-ocean-dns-zone) -- hosts that CNAME when DigitalOcean serves your domain's DNS
