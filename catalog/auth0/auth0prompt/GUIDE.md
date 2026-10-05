# Auth0 Prompt Guide

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

## Prompt Security Notes

### One Flow for Everyone

The prompt settings are tenant-wide: every application that signs people in through the tenant gets the same flow. Apply a change on a pre-production tenant first and walk through a sign-in, a sign-up and a password reset before promoting it.

### Identifier First Tells Who Has an Account Elsewhere

Routing by email domain reveals which domains belong to an enterprise connection -- anyone can type an address and see where it goes. That is how home realm discovery works; keep sensitive customer names out of connection domains you would not publish.

### Device Biometrics Are Offered, Not Required

`webauthnPlatformFirstFactor` offers a stronger first factor; it never removes the password. Require multi-factor authentication in the tenant's multi-factor policy when a password alone is not enough.

## Pairing and Apply Order

### Universal Login Before the Page's Look

Apply this component before the ones that style the login page: Auth0 Branding (logo, colors, page template), Auth0 Prompt Custom Text (the words) and Auth0 Prompt Screen Partials (extra fields) all apply only to the `new` experience. They deploy on `classic` too, and nobody sees them.

### The Biometrics Factor Comes First

Auth0 refuses `webauthnPlatformFirstFactor: true` until WebAuthn with device biometrics is enabled as a multi-factor method on the tenant. Enable the factor (Dashboard: Security, Multi-factor Auth) before this component, and keep it enabled for as long as the setting is on.

### Identifier First Needs Domains to Route

Identifier-first login alone moves the password to a second screen. Routing people to their company's identity provider needs each enterprise connection's identity provider domains set on the connection.

### Destroy Is Not Undo

Destroying this component stops managing the settings and leaves the last-applied values on the tenant. To go back to the Classic pages or to a single login screen, apply that value first, then remove the resource.

### One Resource per Tenant

Two Auth0 Prompt resources on one tenant write the same settings and undo each other on every apply. Keep one per tenant, per provider connection.

## Permissions
## Management API Scopes

Auth0 prompt settings require the following Management API scopes. They must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read | `read:prompts` | Read the tenant's prompt settings |
| Update | `update:prompts` | Change the tenant's prompt settings |

There is no create or delete scope to grant: the prompt settings exist with the tenant and have no delete.

## Compliance
## Prompt Compliance Notes

### The Login Flow as Auditable Configuration

How people sign in -- the experience, the order of the screens, the first factor offered -- is version-controlled with the rest of the tenant's identity configuration, so a change to the login flow is reviewed like any other change. Changes are also recorded in Auth0's tenant logs.

## Cost
## Pricing Model

Auth0 pricing is based on the plan tier and monthly active users, not on settings. Choosing the Universal Login experience and turning on identifier-first login are available on every plan, the Free plan included. WebAuthn with device biometrics is a multi-factor method whose availability depends on the plan.

## Cost Impact

Changing these settings has no billing impact.
