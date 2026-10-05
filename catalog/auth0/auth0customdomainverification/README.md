# Auth0CustomDomainVerification

Completes an [Auth0 custom domain](https://auth0.com/docs/customize/custom-domains): asks Auth0 to check the domain's DNS record and waits until the domain is ready to serve sign-in -- verified, and for an Auth0-managed domain, holding its issued certificate.

## When to Use

- **Bring a sign-in domain live in one install**: the Auth0CustomDomain, the DNS record it answers with, and this verification, ordered after the record.
- **A self-managed domain's proxy key**: Auth0 returns the key your proxy must send, once, at verification.
- **A default domain Auth0 has verified**: `Auth0TenantSettings.default_custom_domain` reads this resource's `domain`.

## Quick Start

```yaml
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

See the Auth0CustomDomain kind for the whole three-resource install.

## Key Behaviors

- **Order it after the record**: Auth0 can only verify a record that already resolves publicly. Declare the DNS record as a `depends_on` relationship so the record is created first.
- **It waits, then fails plainly**: the deploy waits until Auth0 reports the domain `ready` (up to five minutes) and otherwise fails naming the last status. A domain left in `pending_verification` almost always means the record does not resolve yet, or sits behind a CDN proxy or CNAME flattening.
- **A one-time action**: there is nothing to update, and destroying this resource leaves the domain verified. Destroy the Auth0CustomDomain to remove the domain; pointing this resource at another domain replaces it.
- **Permissions**: the credential needs `create:custom_domains` (Auth0 files verification under create) and `read:custom_domains` (`iac/permissions.yaml`).

## Outputs

| Output | Description |
|---|---|
| `custom_domain_id` | The verified custom domain's identifier (`cd_...`) |
| `domain` | The verified domain's name |
| `origin_domain_name` | The tenant host the domain serves from |
| `cname_api_key` | The key a self-managed domain's proxy sends in the `cname-api-key` header (secret; empty for Auth0-managed) |

## Auth0 Documentation

- [Auth0-managed certificates](https://auth0.com/docs/customize/custom-domains/auth0-managed-certificates)
- [Terraform auth0_custom_domain_verification](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/custom_domain_verification)
- [Pulumi auth0.CustomDomainVerification](https://www.pulumi.com/registry/packages/auth0/api-docs/customdomainverification/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
