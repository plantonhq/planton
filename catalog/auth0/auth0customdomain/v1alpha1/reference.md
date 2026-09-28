# Auth0CustomDomain

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `auth0.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

Auth0CustomDomainSpec creates a custom domain for the Auth0 tenant the provider
connection's credential belongs to: a name you own (id.example.com) that serves
the tenant's Universal Login, its Authentication API and the links in its
emails, in place of the tenant's canonical domain (example.eu.auth0.com).

Creating the domain does not make it live. Auth0 answers with a DNS record that
proves you control the name (outputs dns_record_name, dns_record_type,
dns_record_value). Publish it with a DNS record kind, then apply an
Auth0CustomDomainVerification, which waits until Auth0 has confirmed the record
and -- for an Auth0-managed domain -- issued the certificate. Until then the
domain answers nothing and the canonical domain keeps serving.

Once live, the domain and the canonical domain both work. A token carries the
issuer ("iss") of the domain that served the request, so an application that
signs people in through the custom domain must trust
"https://<domain>/" as the issuer. Which domain Auth0 uses for email links is
the tenant's default domain (Auth0TenantSettings.default_custom_domain).

Plans: the Free plan includes one custom domain with Auth0-managed
certificates (Auth0 asks for a card on file, which is not charged). Self-managed
certificates, and more than one custom domain per tenant, need the Enterprise
plan.

The credential needs create:custom_domains, read:custom_domains,
update:custom_domains and delete:custom_domains on the tenant's Management API
(iac/permissions.yaml).

https://auth0.com/docs/customize/custom-domains
https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/custom_domain
https://www.pulumi.com/registry/packages/auth0/api-docs/customdomain/

## Example

```yaml
# Auth0 Custom Domain Test Manifest
# This file is used for testing the Auth0CustomDomain component.
#
# Prerequisites:
# 1. Set the following environment variables:
#    - AUTH0_DOMAIN: the test tenant's domain (e.g., your-test-tenant.auth0.com)
#    - AUTH0_CLIENT_ID: M2M application client ID
#    - AUTH0_CLIENT_SECRET: M2M application client secret
#
# 2. The M2M application must have these scopes:
#    - create:custom_domains
#    - read:custom_domains
#    - update:custom_domains
#    - delete:custom_domains
#
# 3. The tenant must allow a custom domain: a card on file on the Free plan
#    (not charged), or a paid plan.

apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0CustomDomain
metadata:
  name: test-sign-in-domain
  org: test-org
  env: development
  labels:
    purpose: testing
spec:
  # The name people sign in on
  domain: id.example.com

  # Auth0 issues and renews the certificate; publish the CNAME it answers with
  type: auth0_managed_certs

  # The only TLS policy Auth0 accepts
  tlsPolicy: recommended
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.domain` | `string` | yes |  |  |
| `spec.type` | `string` | yes |  |  |
| `spec.customClientIpHeader` | `string` |  |  |  |
| `spec.tlsPolicy` | `string` |  |  |  |
| `spec.domainMetadata` | `map<string, string>` |  |  |  |
| `spec.relyingPartyIdentifier` | `string` |  |  |  |

## Field Details

### spec.domain

`string` · required

domain is the name to serve the tenant on, a host name you control in DNS
(e.g. "id.example.com" or "login.example.com"). Changing it replaces the
custom domain, and the new name must be verified again.

- rule: {"required":true,"string":{"hostname":true}}

### spec.type

`string` · required

type chooses who holds the domain's TLS certificate. Changing it replaces
the custom domain.
- "auth0_managed_certs": Auth0 issues and renews the certificate. You publish
  one CNAME record pointing the domain at the tenant's edge, and keep it:
  renewals (every three months) re-check it. The record must stay DNS-only --
  no CDN proxying and no CNAME flattening in front of it.
- "self_managed_certs": your own reverse proxy terminates TLS with your
  certificate and forwards to the tenant's origin (output
  origin_domain_name), sending the header Auth0 gives it at verification
  (Auth0CustomDomainVerification's cname_api_key). Needs the Enterprise
  plan; control is proven with a TXT record.

- rule: {"required":true,"string":{"in":["auth0_managed_certs","self_managed_certs"]}}

### spec.customClientIpHeader

`string`

custom_client_ip_header names the request header Auth0 reads the end
user's IP address from, for a proxy of yours in front of the domain (IP
allow lists, rate limits and anomaly detection then see the person, not the
proxy). Unset, Auth0 reads x-forwarded-for.
- "x-forwarded-for": the common proxy header, and Auth0's default
- "cf-connecting-ip": Cloudflare's
- "x-azure-clientip": Azure Front Door's
- "true-client-ip": the header some CDNs set instead

- rule: {"string":{"in":["","cf-connecting-ip","x-forwarded-for","true-client-ip","x-azure-clientip"]}}

### spec.tlsPolicy

`string`

tls_policy is the TLS policy Auth0 applies to an Auth0-managed domain.
"recommended" (TLS 1.2 and 1.3 with modern ciphers) is the only policy
Auth0 accepts; the older "compatible" policy is retired. Unset, Auth0
applies "recommended".

- rule: {"string":{"in":["","recommended"]}}

### spec.domainMetadata

`map<string, string>`

domain_metadata is up to ten key-value pairs stored with the domain,
available to Universal Login templates as custom_domain.domain_metadata.<key>
-- for example a brand name, when several domains of one tenant sign in for
different brands. Values are at most 255 characters. Auth0 offers it to
early-access tenants.

- rule: {"map":{"maxPairs":"10","values":{"string":{"maxLen":"255"}}}}

### spec.relyingPartyIdentifier

`string`

relying_party_identifier is the relying-party ID passkeys are bound to on
this domain. A passkey is usable only on its relying party and the
relying party's subdomains, so setting it to a parent domain (e.g.
"example.com" for "id.example.com") lets one passkey sign in across your
domains; Auth0 recommends the parent domain for exactly this. Unset,
passkeys bind to the custom domain itself.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"hostname":true}}

## Validation Rules

- `spec.tls_policy_only_for_auth0_managed_certs`: tls_policy applies only to an Auth0-managed domain (type auth0_managed_certs): with self_managed_certs your own proxy terminates TLS and sets its policy

## Outputs

Reference an output from another manifest as `valueFrom: {kind: Auth0CustomDomain, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the custom domain's identifier in Auth0 (cd_...). An Auth0CustomDomainVerification's custom_domain_id reads it. |
| `status.outputs.domain` | `string` | domain is the custom domain's name, as created (e.g. "id.example.com"). |
| `status.outputs.status` | `string` | status is where the domain is in its life: "pending_verification" until Auth0 has confirmed the DNS record, then "ready" once its certificate is issued ("disabled", "pending" and "failed" are the other states Auth0 reports). |
| `status.outputs.origin_domain_name` | `string` | origin_domain_name is the tenant host the domain serves from (e.g. "example-cd-abc123.edge.tenants.eu.auth0.com"): the CNAME target for an Auth0-managed domain, and the origin a self-managed proxy forwards to. |
| `status.outputs.dns_record_name` | `string` | The DNS record that proves control of the domain -- compose these three into a DNS record kind (for example a CloudflareDnsRecord's name, type and content) to complete verification. For an Auth0-managed domain it is the CNAME the domain is served through, and it must stay in place: Auth0 re-checks it on every certificate renewal. For a self-managed domain it is a TXT record. dns_record_name is the record's fully qualified name (the domain itself for a CNAME, e.g. "id.example.com"; for a TXT record the name Auth0 assigns, e.g. "_cf-custom-hostname.id.example.com"). |
| `status.outputs.dns_record_type` | `string` | dns_record_type is the record's type: "CNAME" or "TXT". |
| `status.outputs.dns_record_value` | `string` | dns_record_value is the record's value: the CNAME's target host, or the TXT record's text. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| Auth0CustomDomainVerification | `spec.customDomainId` | `status.outputs.id` |

## See Also

- [Overview](../README.md)
