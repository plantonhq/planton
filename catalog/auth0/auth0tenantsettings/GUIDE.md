# Auth0 Tenant Settings Guide

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

## Tenant-Settings Security Notes

### Public by Design

Everything this component sets is shown to anyone who opens the tenant's login page: the name, the logo, and the support contacts. Put nothing in them you would not publish.

### The Logo Is Fetched From Where You Host It

The login page loads `pictureUrl` from its host on every visit. Host the logo somewhere you control, over HTTPS, and keep the URL answering: a moved or deleted image breaks the page's look (never the sign-in itself).

### A Support Address Is an Invitation

People who can't sign in will write to `supportEmail`. Point it at a monitored mailbox that knows it must never ask for a password.

## Permissions
## Management API Scopes

Auth0 tenant settings require the following Management API scopes. They must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read | `read:tenant_settings` | Read the tenant's current settings |
| Update | `update:tenant_settings` | Change the tenant's settings |
| Read default domain | `read:custom_domains` | Read the tenant's default domain (only when `defaultCustomDomain` is set) |
| Set default domain | `update:custom_domains` | Set the tenant's default domain (only when `defaultCustomDomain` is set) |

There is no create or delete scope to grant: a tenant can't be created or deleted through the Management API, and its settings have no delete.

## Compliance
## Tenant-Settings Compliance Notes

### Presentation as Auditable Configuration

The tenant's face (its name, logo, and support contacts) is version-controlled with the rest of its identity configuration, so a change to what people see at sign-in is reviewed like any other change. Changes are also recorded in Auth0's tenant logs.

## Cost
## Pricing Model

Auth0 pricing is based on the plan tier and monthly active users, not on settings. The friendly name, logo, and support contacts are editable on every plan, the Free plan included. Customization beyond them is gated on a custom domain rather than a plan: Universal Login page templates need one (Auth0's Free plan includes one custom domain), and this component doesn't touch them.

## Cost Impact

Changing these settings has no billing impact.
