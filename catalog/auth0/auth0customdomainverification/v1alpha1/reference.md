# Auth0CustomDomainVerification

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `auth0.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

Auth0CustomDomainVerificationSpec verifies an Auth0CustomDomain: Auth0 checks
the domain's DNS record and, for an Auth0-managed domain, issues its
certificate. The deploy waits until the domain is ready and fails if it is not
ready within the wait, naming the status Auth0 last reported. A domain left in
pending_verification almost always means its record does not resolve publicly
yet, or resolves through a CDN proxy or CNAME flattening that hides it.

Publish the domain's record first -- the Auth0CustomDomain's dns_record_name,
dns_record_type and dns_record_value, DNS-only -- and order this resource after
it with metadata.relationships (type depends_on): Auth0 can verify only a
record that already resolves.

Verification is a one-time action on the domain, so there is nothing to
update, and destroying this resource leaves the domain verified (destroy the
Auth0CustomDomain to remove the domain itself). Pointing it at another custom
domain replaces it, which verifies that domain.

The credential needs read:custom_domains and create:custom_domains on the
tenant's Management API -- Auth0 files verification under create
(iac/permissions.yaml).

https://auth0.com/docs/customize/custom-domains/auth0-managed-certificates
https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/custom_domain_verification
https://www.pulumi.com/registry/packages/auth0/api-docs/customdomainverification/

## Example

```yaml
# Auth0 Custom Domain Verification Test Manifest
# This file is used for testing the Auth0CustomDomainVerification component.
#
# Prerequisites:
# 1. Set the following environment variables:
#    - AUTH0_DOMAIN: the test tenant's domain (e.g., your-test-tenant.auth0.com)
#    - AUTH0_CLIENT_ID: M2M application client ID
#    - AUTH0_CLIENT_SECRET: M2M application client secret
#
# 2. The M2M application must have these scopes:
#    - create:custom_domains (Auth0 files verification under create)
#    - read:custom_domains
#
# 3. The custom domain's DNS record (its Auth0CustomDomain's dns_record_name,
#    dns_record_type and dns_record_value) must already resolve publicly.

apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0CustomDomainVerification
metadata:
  name: test-sign-in-domain-verification
  org: test-org
  env: development
  labels:
    purpose: testing
  # Verify only after the record that proves control of the domain exists
  relationships:
    - kind: CloudflareDnsRecord
      name: test-sign-in-domain-cname
      type: depends_on
spec:
  # The custom domain to verify, read from the Auth0CustomDomain that created it
  customDomainId:
    valueFrom:
      kind: Auth0CustomDomain
      name: test-sign-in-domain
      fieldPath: status.outputs.id
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.customDomainId` | `string \| valueFrom` | yes |  | Auth0CustomDomain (`status.outputs.id`) |

## Field Details

### spec.customDomainId

`string | valueFrom` · required

custom_domain_id is the custom domain to verify: its Auth0 identifier
(cd_...), or a reference to the Auth0CustomDomain that created it, which
resolves to its status.outputs.id.

- references: Auth0CustomDomain (`status.outputs.id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: Auth0CustomDomain, name: <that resource's name>, fieldPath: status.outputs.id}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: Auth0CustomDomainVerification, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.custom_domain_id` | `string` | custom_domain_id is the verified custom domain's identifier in Auth0 (cd_...). |
| `status.outputs.domain` | `string` | domain is the verified domain's name (e.g. "id.example.com"), ready to serve sign-in. Auth0TenantSettings.default_custom_domain reads it, so a tenant's default domain is always one Auth0 has verified. |
| `status.outputs.origin_domain_name` | `string` | origin_domain_name is the tenant host the domain serves from (e.g. "example-cd-abc123.edge.tenants.eu.auth0.com"), the origin a self-managed proxy forwards to. |
| `status.outputs.cname_api_key` | `string` (sensitive) | cname_api_key is the key a self-managed domain's proxy must send to the origin in the cname-api-key header, so Auth0 accepts its requests. Auth0 returns it once, when the domain is first verified; it is empty for an Auth0-managed domain. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.customDomainId` | Auth0CustomDomain | `status.outputs.id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| Auth0TenantSettings | `spec.defaultCustomDomain` | `status.outputs.domain` |

## See Also

- [Overview](../README.md)
