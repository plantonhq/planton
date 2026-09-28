# Auth0 Email Template Guide

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

## Email-Template Security Notes

### Escape What People Typed

`user.name`, `user.nickname`, `user.given_name`, `user.family_name` and `user.picture` are supplied by the person signing up, and a body that prints them raw lets someone put HTML in your email. Escape them (`{{ user.name | escape }}`); `user.email` and the links Auth0 generates are trusted.

### A Link Acts for the Person

The verification and reset links in these emails act for whoever clicks them. Keep `urlLifetimeInSeconds` short where the link grants access -- an hour for a password reset -- and say in the body how long it works.

### A Fixed Redirect Beats a Templated One

`resultUrl` is where people land after acting on the link. Point it at a fixed page of your own. Auth0 restricted custom redirects on non-enterprise tenants created on or after May 5, 2026 to close an open-redirect abuse path.

## Permissions
## Management API Scopes

Auth0 email templates require the following Management API scopes. They must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read | `read:email_templates` | Read the template (also checked before every create) |
| Create | `create:email_templates` | Create a template the tenant has never customized |
| Update | `update:email_templates` | Change the template, take over an existing one, and disable it on destroy |

There is no delete scope to grant: Auth0 cannot delete a template.

## Judgment
## Pairing With Other Auth0 Kinds

- **Auth0 Email Provider** comes first: Auth0 refuses custom templates on a tenant that sends through its built-in provider, and the template's `from` must be on a domain the provider may send for. The registry declares the provider as this kind's prerequisite.
- **Auth0 Tenant Settings** holds the friendly name and support address a body can print (`{{ friendly_name }}`, `{{ support_email }}`), so one change there reaches every email.
- **Auth0 Connection** decides which emails are sent at all: a database connection sends verification and reset emails, a passwordless email connection sends its codes.
- **Auth0 Action** on the `password-reset-post-challenge` trigger chooses where people land after a reset, which the reset template's `resultUrl` cannot under Universal Login.

## The Order to Apply

1. The email provider (Auth0 Email Provider), with its sending domain verified.
2. The tenant settings the bodies print.
3. The templates, one per email.

## Traps

- **No provider, no template**: on a tenant without its own email provider, Auth0 refuses the template.
- **Destroy does not bring back an empty slot**: the template is disabled, not deleted, and the tenant sends Auth0's default email. Its last content stays stored in Auth0, and the next apply of the same template takes it over again.
- **`resultUrl` can be refused**: non-enterprise tenants created on or after May 5, 2026 answer a custom `resultUrl` with 403. Leave it unset there.
- **Universal Login ignores the reset redirect**: after a password reset, Universal Login sends people to the default login route whatever `resultUrl` says; use the `password-reset-post-challenge` Action trigger instead.
- **A sender on `@auth0.com` is ignored**: a `from` containing `@auth0.com` makes the tenant send only default emails.
- **HTML only**: templates do not support plain-text emails.
- **Only some templates carry a link**: `resultUrl` and `urlLifetimeInSeconds` apply to the link templates (verification, reset, blocked account); a welcome email has no link to time or redirect.
- **One template per email**: two resources naming the same `template` overwrite each other.

## Compliance
## Email-Template Compliance Notes

### What People Are Told as Auditable Configuration

The words every verification, reset and security email says are version-controlled with the rest of the tenant's identity configuration, so a change to what people are told about their account is reviewed like any other change. Changes are also recorded in Auth0's tenant logs.

## Cost
## Pricing Model

Auth0 publishes no price for email templates; every plan, the Free plan included, can customize them once the tenant sends through its own email provider. Auth0 pricing is based on the plan tier and monthly active users. Custom `resultUrl` values need an enterprise plan on tenants created on or after May 5, 2026.

## Cost Impact

A template changes what an email says, not how many are sent. The sending service of the tenant's email provider bills its own volume.
