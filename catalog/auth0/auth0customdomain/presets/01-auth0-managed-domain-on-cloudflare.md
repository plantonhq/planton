# Auth0-Managed Domain on Cloudflare

This preset serves the tenant's sign-in on your own domain with a certificate Auth0 issues and renews. It is the first of three resources that install together: this custom domain, a Cloudflare DNS Record whose content is the domain's `dns_record_value`, and an Auth0 Custom Domain Verification ordered after that record. It is the configuration almost every tenant wants, and the one the Free plan includes.

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareDnsRecord
metadata:
  name: sign-in-domain-cname
spec:
  zoneId:
    value: <cloudflare-zone-id>
  name: id.example.com
  type: CNAME
  content:
    valueFrom:
      kind: Auth0CustomDomain
      name: sign-in-domain
      fieldPath: status.outputs.dns_record_value
  proxied: false
---
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0CustomDomainVerification
metadata:
  name: sign-in-domain-verification
  relationships:
    - kind: CloudflareDnsRecord
      name: sign-in-domain-cname
      type: depends_on
spec:
  customDomainId:
    valueFrom:
      kind: Auth0CustomDomain
      name: sign-in-domain
      fieldPath: status.outputs.id
```

## When to Use

- Any tenant real people sign in through, on the Free plan or above
- The first step before branding the Universal Login page, which needs a custom domain
- Tenants whose DNS is on Cloudflare (any DNS kind can publish the same record)

## Key Configuration Choices

- **Auth0-managed certificate** (`type: auth0_managed_certs`) -- Auth0 issues and renews it; nothing to rotate yourself
- **Recommended TLS** (`tlsPolicy: recommended`) -- TLS 1.2 and 1.3, the only policy Auth0 accepts
- **DNS-only CNAME** (`proxied: false` on the record) -- a CDN proxy or CNAME flattening hides the target from Auth0; keep the record for the domain's life, as every renewal re-checks it
- **One install** -- the record reads the domain's output and the verification depends on the record, so nothing is copied by hand

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.domain` | The domain to sign in on (sample: `id.example.com`) | A subdomain of a zone you control |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-passkeys-across-your-domains** -- the same domain with passkeys bound to the parent domain
- **03-self-managed-certificate** -- your own proxy and certificate in front (Enterprise plan)
