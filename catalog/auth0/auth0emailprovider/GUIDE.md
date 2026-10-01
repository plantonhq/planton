# Auth0 Email Provider Guide

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

## Email-Provider Security Notes

### The Credential Sends as You

Whoever holds the provider's credential can send email that passes as your domain's. Give it the least the service allows -- a Resend or SendGrid key with sending access only, an IAM user allowed only `ses:SendRawEmail` -- keep it in an organization secret, and rotate it at the service, then apply again.

### Publish SPF and DKIM for the Sender

`defaultFromAddress` (and every template's `from`) must be on a domain verified at the service, with its SPF and DKIM records published. Without them, emails are dropped, filtered as spam, or shown "on behalf of" the service.

### Verification and Reset Links Travel in These Emails

A password-reset or verification email carries a link that acts for the person. A misconfigured provider fails closed (the email does not arrive), but a provider pointed at a relay you do not control hands those links to it. Point the provider only at services you operate or contract.

### Prefer an Integration to SMTP for Microsoft 365

Auth0's SMTP support uses the LOGIN protocol, and Microsoft is retiring basic authentication on Exchange Online's SMTP endpoints. Use the `ms365` arm (an app registration granted `Mail.Send`) rather than `smtp` against `smtp.office365.com`.

## Permissions
## Management API Scopes

Auth0 email providers require the following Management API scopes. They must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read | `read:email_provider` | Read the tenant's provider (also checked before every create) |
| Create | `create:email_provider` | Set up a provider on a tenant that has none |
| Update | `update:email_provider` | Change the provider, or take over one that already exists |
| Delete | `delete:email_provider` | Delete the provider on destroy |

## Judgment
## Pairing With Other Auth0 Kinds

- **Auth0 Email Template** needs this provider: Auth0 refuses custom templates on a tenant that sends through its built-in provider. Apply the provider first, then the templates.
- **Auth0 Action** is the sender for the `custom` arm. Deploy the Action on the `custom-email-provider` trigger and bind it before setting the provider, so something can send the moment the provider switches; a tenant allows one Action on that trigger.
- **Auth0 Tenant Settings** holds the tenant's friendly name and support address, which Auth0's default emails and your templates (`{{ friendly_name }}`, `{{ support_email }}`) show.
- **Auth0 Custom Domain** decides where the links in the emails point, once it is the tenant's default domain.

## The Order to Apply

1. The sending domain, verified at the service (SPF and DKIM published).
2. The organization secrets holding the service's credentials.
3. This provider.
4. The templates.

## Traps

- **Resend is not a name the provider accepts**: Auth0's dashboard and API offer Resend directly, but terraform-provider-auth0 v1.58.0 rejects `resend` (`internal/auth0/email/resource.go:35-38`, the `name` validation). Use the `smtp` arm with `smtp.resend.com`, user `resend`, port 587, and an API key as the password.
- **Applying takes over the tenant's provider**: a tenant has one provider, and the create rewrites whichever one exists. Two Auth0EmailProvider resources on one tenant fight each other.
- **Destroy switches the tenant back to the built-in provider**: it sends at most 10 emails per minute, from `no-reply@auth0user.net`, and ignores custom templates -- a surprise for people waiting on a reset email.
- **Changing the SMTP host, port or user resends the password**: Auth0 requires the password with any of them; the modules send the whole credentials block together whenever it changes.
- **Auth0 connects from its own IP addresses**: an SMTP server that filters by source must allow Auth0's published outbound addresses.
- **Test domains are refused**: Auth0 rejects some placeholder domains commonly used for testing; use a real, verified sender.

## Compliance
## Email-Provider Compliance Notes

### The Sending Path as Auditable Configuration

Which service carries the tenant's emails, and from which address, is version-controlled with the rest of the tenant's identity configuration, and its credentials live in the organization's secret store. Changes are also recorded in Auth0's tenant logs, and each service keeps its own delivery log.

## Cost
## Pricing Model

Auth0 publishes no price for an email provider and gates it on no plan: every tenant, the Free plan included, can send through its own service. Auth0 pricing is based on the plan tier and monthly active users.

## Cost Impact

The sending service bills its own volume on its own plan (Resend, Amazon SES, SendGrid, and the rest). The emails Auth0 sends -- one per verification, reset, invitation or email code -- scale with sign-ups and sign-ins.
