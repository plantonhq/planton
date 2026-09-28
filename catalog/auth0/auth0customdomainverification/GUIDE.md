# Auth0 Custom Domain Verification Guide

## Security
## Platform Security Posture

The certifications below are Auth0's own published claims about their hosted platform (verify current status on Auth0's compliance page). They describe the vendor's service — never this Planton component, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

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

## Verification Security Notes

### Verification Proves DNS Control, Nothing More

Auth0 verifies a domain by finding its record in public DNS. Whoever can publish that record can bring the domain live for your tenant, so protect the zone's credentials as you protect the tenant's.

### The Proxy Key Is a Credential

For a self-managed domain, `cname_api_key` authenticates your proxy to Auth0's origin. Auth0 returns it only once, at the first successful verification; Planton keeps it as a secret output. Hand it to the proxy through your secret store, never a plain configuration file.

## Permissions
## Management API Scopes

Verifying a custom domain requires the following Management API scopes. They must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Verify | `create:custom_domains` | Start the verification (Auth0 files it under create) |
| Read | `read:custom_domains` | Poll the domain until it is ready, and read it back |

## Compliance
## Verification Compliance Notes

### A Reviewed Step to Going Live

Bringing a sign-in domain live is a declared, version-controlled step, ordered after the DNS change that makes it possible, so the moment a new sign-in address starts serving is reviewed like any other change. Verifications are also recorded in Auth0's tenant logs.

## Cost
## Pricing Model

Verification is free: Auth0 checks the record and, for an Auth0-managed domain, issues and renews its certificate at no charge on every plan that allows the custom domain itself.

## Cost Impact

None beyond the custom domain's own plan requirements.
