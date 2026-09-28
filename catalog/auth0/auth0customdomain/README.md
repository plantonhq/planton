# Auth0CustomDomain

Serves an [Auth0 tenant](https://auth0.com/docs/customize/custom-domains) on a domain you own -- `id.example.com` instead of `example.eu.auth0.com` -- so Universal Login, the links in the tenant's emails, and the issuer of its tokens all carry your name.

## When to Use

- **Your name in the address bar**: people sign in on `id.example.com`, not on an `auth0.com` address they have never heard of.
- **Branded pages**: a Universal Login page template (the `auth0_branding` page body) needs a custom domain first.
- **Passkeys across your domains**: bind passkeys to a parent domain so one passkey signs in everywhere you serve.
- **Infrastructure as code**: the domain, the DNS record that proves it, and its verification are one install.

## Quick Start

Three resources, one install: the domain, the record Auth0 answers with, and the verification that waits for it.

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0CustomDomain
metadata:
  name: sign-in-domain
spec:
  domain: id.example.com
  type: auth0_managed_certs
---
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

## Key Behaviors

- **Created is not live**: Auth0 answers a new domain with the DNS record that proves you control it (`dns_record_name`, `dns_record_type`, `dns_record_value`). The domain serves nothing until an `Auth0CustomDomainVerification` has confirmed that record.
- **Two ways to hold the certificate**: `auth0_managed_certs` (Auth0 issues and renews it; you keep one DNS-only CNAME) or `self_managed_certs` (your own proxy terminates TLS; Enterprise plan).
- **Both domains keep working**: the tenant's canonical domain still answers. A token carries the issuer of the domain that served the request, so an application that signs in through the custom domain trusts `https://<domain>/`.
- **Email links** use the tenant's default domain: set it with `Auth0TenantSettings.default_custom_domain`.
- **Replacing**: changing `domain` or `type` replaces the custom domain, and the new one is verified again. Destroy deletes it.
- **Plans**: the Free plan includes one custom domain with Auth0-managed certificates (a card on file, not charged). Self-managed certificates and more than one domain need the Enterprise plan.
- **Permissions**: the credential needs `create:custom_domains`, `read:custom_domains`, `update:custom_domains` and `delete:custom_domains` (`iac/permissions.yaml`).

## Outputs

| Output | Description |
|---|---|
| `id` | The custom domain's identifier in Auth0 (`cd_...`) |
| `domain` | The custom domain's name |
| `status` | `pending_verification` until verified, then `ready` |
| `origin_domain_name` | The tenant host the domain serves from |
| `dns_record_name` | The name of the record that proves control (the domain itself for a CNAME) |
| `dns_record_type` | `CNAME` (Auth0-managed) or `TXT` (self-managed) |
| `dns_record_value` | The record's value: the CNAME target, or the TXT text |

## Auth0 Documentation

- [Custom domains](https://auth0.com/docs/customize/custom-domains)
- [Terraform auth0_custom_domain](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/custom_domain)
- [Pulumi auth0.CustomDomain](https://www.pulumi.com/registry/packages/auth0/api-docs/customdomain/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
