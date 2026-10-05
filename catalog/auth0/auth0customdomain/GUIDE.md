# Auth0 Custom Domain Guide

## Security
## Platform Security Posture

The certifications below are Auth0's own published claims about their hosted platform (verify current status on Auth0's compliance page). They describe the vendor's service — never this catalog kind, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Auth0's published certifications and security standards:

- SOC 2 Type II (annual audit)
- ISO 27001, ISO 27018 (privacy controls)
- HIPAA BAA available on enterprise plans
- PCI DSS Level 1 Service Provider
- FedRAMP Authorized (moderate baseline)
- CSA STAR Level 2
- GDPR compliant with Data Processing Agreement

## Data Protection

- **Data residency**: US, EU, AU regions
- **Encryption in transit**: TLS 1.2+
- **Encryption at rest**: AES-256
- **Penetration testing**: Annual third-party assessments

## Custom-Domain Security Notes

### Keep the CNAME DNS-Only, and Keep It

An Auth0-managed domain is served through its CNAME, and Auth0 re-checks the CNAME every time it renews the certificate (about every three months). A CDN proxy or CNAME flattening in front of it hides the target: verification fails, and a later renewal can lapse. Leave the record DNS-only for as long as the domain exists.

### Whoever Controls the DNS Controls the Sign-In

The domain is proven, and its certificate issued, by whoever can edit its DNS. Protect the zone as you protect the tenant: least-privilege DNS credentials, and review for changes to this record.

### The Issuer Changes With the Domain

A token requested through the custom domain carries `https://<domain>/` as its issuer. Point an application's issuer setting at the domain it signs people in through; trusting both the canonical and the custom issuer is only needed while applications move over.

### A Self-Managed Proxy Holds a Key

With `self_managed_certs`, Auth0 accepts requests from your proxy only when they carry the `cname-api-key` header, returned once by the verification (Auth0 Custom Domain Verification's `cname_api_key`, a secret output). Store it like any credential.

## Permissions
## Management API Scopes

Auth0 custom domains require the following Management API scopes. They must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Create | `create:custom_domains` | Create the custom domain |
| Read | `read:custom_domains` | Read the domain and its verification record |
| Update | `update:custom_domains` | Change its TLS policy, client-IP header, metadata, or relying party |
| Delete | `delete:custom_domains` | Delete the custom domain |

## Compliance
## Custom-Domain Compliance Notes

### The Sign-In Address as Auditable Configuration

The address people sign in on, its certificate type, and its TLS policy are version-controlled with the rest of the tenant's identity configuration, so a change to where people sign in is reviewed like any other change. Changes are also recorded in Auth0's tenant logs.

## Cost
## Pricing Model

Every Auth0 plan, the Free plan included, has one custom domain with Auth0-managed certificates at no charge; on the Free plan Auth0 asks for a card on file first and does not charge it. Self-managed certificates, and more than one custom domain per tenant, need the Enterprise plan.

## Cost Impact

The custom domain changes where people sign in, not what a sign-in costs. A self-managed domain adds whatever your own proxy and certificate cost.
